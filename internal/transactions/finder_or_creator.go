package transactions

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"ulascansenturk/service/internal/accounts"
	"ulascansenturk/service/internal/constants"
)

type FinderOrCreator interface {
	Call(ctx context.Context, params *DBTransaction) (*Transaction, error)
}

type FinderOrCreatorService struct {
	transactionRepo Repository
	accountRepo     accounts.Repository
	validate        *validator.Validate
}

func NewFinderOrCreatorService(transactionRepo Repository, accountRepo accounts.Repository, validate *validator.Validate) *FinderOrCreatorService {
	return &FinderOrCreatorService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		validate:        validate,
	}
}

func (s *FinderOrCreatorService) Call(ctx context.Context, params *DBTransaction) (*Transaction, error) {
	existingTransaction, err := s.transactionRepo.GetByReferenceID(ctx, params.ReferenceID)
	if err != nil {
		return nil, err
	}

	if existingTransaction != nil {
		return existingTransaction, nil
	}

	account, accountErr := s.accountRepo.GetByID(ctx, params.AccountID)
	if accountErr != nil {
		return nil, accountErr
	}

	if account == nil {
		return nil, fmt.Errorf("account is missing: %s", params.AccountID)
	}

	if params.UserID == nil {
		params.UserID = &account.UserID
	}

	switch account.Status {
	case constants.AccountStatusACTIVE:
		return s.createTransaction(ctx, params)
	case constants.AccountStatusCLOSED:
		return s.createFailedTransaction(ctx, params, account)
	default:
		return nil, fmt.Errorf("account is not ACTIVE: %s, status: %s ", account.ID, account.Status)
	}
}

func (s *FinderOrCreatorService) validAccounts(sourceAccount, destinationAccount *accounts.Account) bool {
	if sourceAccount.Status != constants.AccountStatusACTIVE && destinationAccount.Status != constants.AccountStatusACTIVE {
		return false
	}

	return true
}

func (s *FinderOrCreatorService) createTransaction(ctx context.Context, params *DBTransaction) (*Transaction, error) {
	createdTransaction, createTransactionErr := s.transactionRepo.Create(ctx, FromDBTransaction(params))
	if createTransactionErr != nil {
		return nil, createTransactionErr
	}

	return createdTransaction, nil
}

func (s *FinderOrCreatorService) createFailedTransaction(ctx context.Context, params *DBTransaction, account *accounts.Account) (*Transaction, error) {
	if params.Status != constants.TransactionStatusFAILURE {
		return nil, fmt.Errorf("only FAILURE transaction can be created for not INACTIVE accounts: %s", account.ID)
	}

	return s.createTransaction(ctx, params)
}
