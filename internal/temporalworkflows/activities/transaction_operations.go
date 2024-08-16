package activities

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"time"
	"ulascansenturk/service/internal/accounts"
	"ulascansenturk/service/internal/constants"
	"ulascansenturk/service/internal/transactions"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
)

type TransactionOperations struct {
	finderOrCreatorService transactions.FinderOrCreator
	transactionService     transactions.Service
	accountsService        accounts.Service
}

func NewTransactionOperations(finderOrCreatorService transactions.FinderOrCreator, transactionsService transactions.Service, accountsService accounts.Service) *TransactionOperations {
	return &TransactionOperations{
		finderOrCreatorService: finderOrCreatorService,
		transactionService:     transactionsService,
		accountsService:        accountsService,
	}
}

type TransferParams struct {
	Amount                            int
	FeeAmount                         *int
	Metadata                          *map[string]interface{}
	DestinationAccountID              uuid.UUID
	SourceTransactionReferenceID      uuid.UUID
	DestinationTransactionReferenceID uuid.UUID
	FeeTransactionReferenceID         uuid.UUID
	SourceAccountID                   uuid.UUID
}

type TransferResult struct {
	SourceTransactionReferenceID      uuid.UUID
	DestinationTransactionReferenceID uuid.UUID
	FeeTransactionReferenceID         uuid.UUID
	FeeTransaction                    *transactions.Transaction
	SourceTransaction                 *transactions.Transaction
	DestinationTransaction            *transactions.Transaction
}

func (t *TransactionOperations) Transfer(ctx context.Context, params TransferParams) (*TransferResult, error) {
	tx, err := t.beginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer t.finalizeTransaction(tx, &err)

	validAccounts, accountsErr := t.validateAccount(ctx, params.Amount, params.FeeAmount, params.SourceAccountID, params.DestinationAccountID)
	if accountsErr != nil {
		return nil, temporal.NewNonRetryableApplicationError("Error on validating accounts", "validate-accounts-err", accountsErr)

	}

	pendingOutGoingTransaction, err := t.createPendingOutgoingTransaction(ctx, params, *validAccounts.SourceAccount)
	if err != nil {
		return nil, err
	}

	pendingFeeTrx, err := t.createPendingFeeTransaction(ctx, params, *validAccounts.SourceAccount)
	if err != nil {
		return nil, err
	}

	pendingIncomingTransaction, err := t.createPendingIncomingTransaction(ctx, params, *validAccounts.DestinationAccount)
	if err != nil {
		return nil, err
	}

	if err := t.updateAccountBalances(ctx, params, tx); err != nil {
		return nil, err
	}

	if err := t.finalizeTransactions(ctx, pendingOutGoingTransaction, pendingIncomingTransaction, pendingFeeTrx, tx); err != nil {
		return nil, err
	}

	return t.createTransferResult(params, pendingOutGoingTransaction, pendingIncomingTransaction, pendingFeeTrx), nil
}

func (t *TransactionOperations) beginTransaction(ctx context.Context) (*gorm.DB, error) {
	tx, err := t.transactionService.BeginTransaction(ctx)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "error while starting database transaction", nil)
	}
	return tx, nil
}

func (t *TransactionOperations) finalizeTransaction(tx *gorm.DB, err *error) {
	if *err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
}

func (t *TransactionOperations) createPendingOutgoingTransaction(ctx context.Context, params TransferParams, sourceAccount accounts.Account) (*transactions.Transaction, error) {
	pendingOutgoingTransactionParams := &transactions.DBTransaction{
		UserID:       &sourceAccount.UserID,
		Amount:       params.Amount,
		AccountID:    sourceAccount.ID,
		CurrencyCode: constants.CurrencyCode(sourceAccount.Currency),
		ReferenceID:  params.SourceTransactionReferenceID,
		Metadata: &map[string]interface{}{
			"OperationType":        "Transfer",
			"LinkedTransactionID":  params.SourceTransactionReferenceID.String(),
			"LinkedAccountID":      sourceAccount.ID.String(),
			"DestinationAccountID": params.DestinationAccountID.String(),
			"timestamp":            time.Now().String(),
		},
		Status:          constants.TransactionStatusPENDING,
		TransactionType: constants.TransactionTypeOUTBOUND,
	}
	pendingOutGoingTransaction, err := t.findOrCreateTransaction(ctx, pendingOutgoingTransactionParams)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "error while creating pending outgoing trx", nil)
	}
	return pendingOutGoingTransaction, nil
}

func (t *TransactionOperations) createPendingFeeTransaction(ctx context.Context, params TransferParams, sourceAccount accounts.Account) (*transactions.Transaction, error) {
	if params.FeeAmount == nil {
		return nil, nil
	}
	pendingOutgoingFeeTransactionParams := &transactions.DBTransaction{
		UserID:       &sourceAccount.UserID,
		Amount:       *params.FeeAmount,
		AccountID:    sourceAccount.ID,
		CurrencyCode: constants.CurrencyCode(sourceAccount.Currency),
		ReferenceID:  params.FeeTransactionReferenceID,
		Metadata: &map[string]interface{}{
			"OperationType":       "Fee Transfer",
			"LinkedTransactionID": params.FeeTransactionReferenceID.String(),
			"LinkedAccountID":     params.SourceAccountID.String(),
			"timestamp":           time.Now().String(),
		},
		Status:          constants.TransactionStatusPENDING,
		TransactionType: constants.TransactionTypeOUTGOINGFEE,
	}
	pendingOutGoingFeeTransaction, err := t.findOrCreateTransaction(ctx, pendingOutgoingFeeTransactionParams)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "error while creating pending outgoing fee trx", nil)
	}
	return pendingOutGoingFeeTransaction, nil
}

func (t *TransactionOperations) createPendingIncomingTransaction(ctx context.Context, params TransferParams, destinationAccount accounts.Account) (*transactions.Transaction, error) {
	pendingIncomingTransactionParams := &transactions.DBTransaction{
		UserID:       &destinationAccount.UserID,
		Amount:       params.Amount,
		AccountID:    destinationAccount.ID,
		CurrencyCode: constants.CurrencyCode(destinationAccount.Currency),
		Metadata: &map[string]interface{}{
			"OperationType":        "Transfer",
			"LinkedTransactionID":  params.DestinationTransactionReferenceID.String(),
			"LinkedAccountID":      params.DestinationAccountID.String(),
			"DestinationAccountID": params.DestinationAccountID.String(),
			"SourceAccountID":      params.SourceAccountID,
			"timestamp":            time.Now().String(),
		},
		ReferenceID:     params.DestinationTransactionReferenceID,
		Status:          constants.TransactionStatusPENDING,
		TransactionType: constants.TransactionTypeINBOUND,
	}
	pendingIncomingTransaction, err := t.findOrCreateTransaction(ctx, pendingIncomingTransactionParams)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "error while creating pending incoming trx", nil)
	}
	return pendingIncomingTransaction, nil
}

func (t *TransactionOperations) updateAccountBalances(ctx context.Context, params TransferParams, tx *gorm.DB) error {
	sourceAccount, err := t.accountsService.GetAccountByID(ctx, params.SourceAccountID)
	if err != nil {
		return err
	}

	sourceAccount.Balance -= params.Amount
	if params.FeeAmount != nil {
		sourceAccount.Balance -= *params.FeeAmount
	}

	if err := t.accountsService.UpdateAccount(ctx, sourceAccount, tx); err != nil {
		return err
	}

	destinationAccount, err := t.accountsService.GetAccountByID(ctx, params.DestinationAccountID)
	if err != nil {
		return err
	}

	destinationAccount.Balance += params.Amount

	return t.accountsService.UpdateAccount(ctx, destinationAccount, tx)
}

func (t *TransactionOperations) finalizeTransactions(ctx context.Context, outgoing, incoming, fee *transactions.Transaction, tx *gorm.DB) error {
	outgoing.Status = constants.TransactionStatusSUCCESS
	if err := t.transactionService.UpdateTransaction(ctx, outgoing, tx); err != nil {
		return err
	}

	incoming.Status = constants.TransactionStatusSUCCESS
	if err := t.transactionService.UpdateTransaction(ctx, incoming, tx); err != nil {
		return err
	}

	if fee != nil {
		fee.Status = constants.TransactionStatusSUCCESS
		if err := t.transactionService.UpdateTransaction(ctx, fee, tx); err != nil {
			return err
		}
	}

	return nil
}

func (t *TransactionOperations) createTransferResult(params TransferParams, outgoing, incoming, fee *transactions.Transaction) *TransferResult {
	return &TransferResult{
		SourceTransactionReferenceID:      params.SourceTransactionReferenceID,
		DestinationTransactionReferenceID: params.DestinationTransactionReferenceID,
		FeeTransactionReferenceID:         params.FeeTransactionReferenceID,
		FeeTransaction:                    fee,
		SourceTransaction:                 outgoing,
		DestinationTransaction:            incoming,
	}
}

func (t *TransactionOperations) findOrCreateTransaction(ctx context.Context, params *transactions.DBTransaction) (*transactions.Transaction, error) {
	transaction, transactionErr := t.finderOrCreatorService.Call(ctx, params)
	if transactionErr != nil {
		return nil, transactionErr
	}

	return transaction, nil
}

func (t *TransactionOperations) validateAccount(ctx context.Context, transferAmount int, feeAmount *int, sourceAccountID uuid.UUID, destinationAccountID uuid.UUID) (*ValidAccounts, error) {
	sourceAccount, accountErr := t.accountsService.GetAccountByID(ctx, sourceAccountID)
	if accountErr != nil {
		return nil, accountErr
	}

	if sourceAccount == nil {
		return nil, fmt.Errorf("account not found: %s", sourceAccountID)
	}

	if sourceAccount.Status != constants.AccountStatusACTIVE {
		return nil, fmt.Errorf("account is not active: %s", sourceAccount.ID)
	}

	destinationAccount, destinationAccountErr := t.accountsService.GetAccountByID(ctx, destinationAccountID)
	if destinationAccountErr != nil {
		return nil, destinationAccountErr
	}

	if destinationAccount == nil {
		return nil, fmt.Errorf("account not found: %s", destinationAccountID)
	}

	if destinationAccount.Status != constants.AccountStatusACTIVE {
		return nil, fmt.Errorf("account is not active: %s", destinationAccount.ID)
	}

	totalAmount := transferAmount
	if feeAmount != nil {
		totalAmount += *feeAmount
	}

	if totalAmount > sourceAccount.Balance {
		return nil, fmt.Errorf("insufficient balance! transfer amount: %d,  account balance: %d", totalAmount, sourceAccount.Balance)
	}

	return &ValidAccounts{
		SourceAccount:      sourceAccount,
		DestinationAccount: destinationAccount,
	}, nil
}

type ValidAccounts struct {
	SourceAccount      *accounts.Account
	DestinationAccount *accounts.Account
}
