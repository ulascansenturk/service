package activities

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
	"time"
	"ulascansenturk/service/internal/accounts"
	accountMocks "ulascansenturk/service/internal/accounts/mocks"
	"ulascansenturk/service/internal/constants"
	"ulascansenturk/service/internal/transactions"
	"ulascansenturk/service/internal/transactions/mocks"

	mockTime "ulascansenturk/service/internal/helpers/mocks"
	"ulascansenturk/service/internal/users"
)

type transactionOperationsSuite struct {
	suite.Suite

	mockGormDB *gorm.DB
	mockDB     sqlmock.Sqlmock

	ctx                    context.Context
	finderOrCreatorService *mocks.MockFinderOrCreator
	transactionsService    *mocks.MockService
	accountsService        *accountMocks.MockService
	timeProvider           *mockTime.MockTimeProvider

	transactionOperations *TransactionOperations
}

func (s *transactionOperationsSuite) SetupTest() {
	s.ctx = context.Background()

	db, mockDB, _ := sqlmock.New()
	s.mockDB = mockDB

	dialector := postgres.New(postgres.Config{Conn: db})
	gormDB, _ := gorm.Open(dialector, &gorm.Config{})

	s.mockGormDB = gormDB

	s.timeProvider = new(mockTime.MockTimeProvider)

	s.finderOrCreatorService = mocks.NewMockFinderOrCreator(s.T())
	s.transactionsService = mocks.NewMockService(s.T())
	s.accountsService = accountMocks.NewMockService(s.T())

	s.transactionOperations = NewTransactionOperations(s.finderOrCreatorService, s.transactionsService, s.accountsService, s.timeProvider)
}

func TestTransactionsOperationsSuite(t *testing.T) {
	suite.Run(t, new(transactionOperationsSuite))
}

func (s *transactionOperationsSuite) TestTransactionOperations() {
	s.Run("Process Tranfer Activity", func() {
		sourceAccID := uuid.New()

		destinationAccID := uuid.New()

		sourceAccountUser := users.User{
			ID:        uuid.New(),
			Email:     "ulas@gmail.com	",
			FirstName: "ulas",
			IsActive:  true,
		}

		sourceAccount := accounts.Account{
			ID:        sourceAccID,
			UserID:    sourceAccountUser.ID,
			Balance:   1000,
			Currency:  "USD",
			Status:    "ACTIVE",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		destinationAccountUser := users.User{
			ID:        uuid.New(),
			Email:     "meric@gmail.com	",
			FirstName: "meric",
			IsActive:  true,
		}

		destinationAccount := accounts.Account{
			ID:        destinationAccID,
			UserID:    destinationAccountUser.ID,
			Balance:   1000,
			Currency:  "USD",
			Status:    "ACTIVE",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		destinationTransactionReference := uuid.New()

		sourceTransactionReference := uuid.New()

		transferParams := TransferParams{
			Amount:                            400,
			DestinationAccountID:              destinationAccount.ID,
			SourceTransactionReferenceID:      sourceTransactionReference,
			DestinationTransactionReferenceID: destinationTransactionReference,
			SourceAccountID:                   sourceAccount.ID,
		}

		s.accountsService.On("GetAccountByID", mock.Anything, sourceAccID).Return(&sourceAccount, nil).Once()

		s.accountsService.On("GetAccountByID", mock.Anything, destinationAccID).Return(&destinationAccount, nil).Once()

		mockTime := time.Date(2024, 8, 17, 20, 13, 19, 662250000, time.UTC)
		s.timeProvider.On("Now").Return(mockTime)

		sourceTransactionMetadata := map[string]interface{}{
			"OperationType":        "Transfer",
			"LinkedTransactionID":  transferParams.SourceTransactionReferenceID.String(),
			"LinkedAccountID":      sourceAccount.ID.String(),
			"DestinationAccountID": transferParams.DestinationAccountID.String(),
			"timestamp":            mockTime,
		}

		destinationTransactionMetadata := map[string]interface{}{
			"OperationType":        "Transfer",
			"LinkedTransactionID":  transferParams.DestinationTransactionReferenceID.String(),
			"LinkedAccountID":      destinationAccount.ID.String(),
			"DestinationAccountID": transferParams.DestinationAccountID.String(),
			"SourceAccountID":      transferParams.SourceAccountID,
			"timestamp":            mockTime,
		}

		pendingOutGoingTransaction := &transactions.Transaction{
			ID:              uuid.New(),
			UserID:          &sourceAccount.UserID,
			Amount:          400,
			AccountID:       sourceAccount.ID,
			CurrencyCode:    constants.CurrencyCode(sourceAccount.Currency),
			ReferenceID:     sourceTransactionReference,
			Metadata:        &sourceTransactionMetadata,
			Status:          constants.TransactionStatusPENDING,
			TransactionType: constants.TransactionTypeOUTBOUND,
		}

		pendingIncomingTransaction := &transactions.Transaction{
			ID:              uuid.New(),
			UserID:          &destinationAccount.UserID,
			Amount:          400,
			AccountID:       destinationAccount.ID,
			CurrencyCode:    constants.CurrencyCode(sourceAccount.Currency),
			ReferenceID:     destinationTransactionReference,
			Metadata:        &destinationTransactionMetadata,
			Status:          constants.TransactionStatusPENDING,
			TransactionType: constants.TransactionTypeINBOUND,
		}

		s.finderOrCreatorService.On("Call", mock.Anything, &transactions.DBTransaction{
			UserID:          &sourceAccount.UserID,
			Amount:          400,
			AccountID:       sourceAccount.ID,
			CurrencyCode:    constants.CurrencyCode(sourceAccount.Currency),
			ReferenceID:     transferParams.SourceTransactionReferenceID,
			Metadata:        &sourceTransactionMetadata,
			Status:          constants.TransactionStatusPENDING,
			TransactionType: constants.TransactionTypeOUTBOUND,
		}).Return(pendingOutGoingTransaction, nil)

		s.finderOrCreatorService.On("Call", mock.Anything, &transactions.DBTransaction{
			UserID:          &destinationAccount.UserID,
			Amount:          400,
			AccountID:       destinationAccount.ID,
			CurrencyCode:    constants.CurrencyCode(sourceAccount.Currency),
			ReferenceID:     transferParams.DestinationTransactionReferenceID,
			Metadata:        &destinationTransactionMetadata,
			Status:          constants.TransactionStatusPENDING,
			TransactionType: constants.TransactionTypeINBOUND,
		}).Return(pendingIncomingTransaction, nil)

		s.accountsService.On("UpdateBalance", mock.Anything, sourceAccID.ID, transferParams.Amount, constants.BalanceOperationDECREASE.String()).Return(nil).Once()

		s.accountsService.On("UpdateBalance", mock.Anything, destinationAccID.ID, transferParams.Amount, constants.BalanceOperationINCREASE.String()).Return(nil).Once()

		s.transactionsService.On("UpdateTransactionStatus", mock.Anything, pendingOutGoingTransaction.ID, constants.TransactionStatusSUCCESS).Once()

		s.transactionsService.On("UpdateTransactionStatus", mock.Anything, pendingIncomingTransaction.ID, constants.TransactionStatusSUCCESS).Once()

		_, err := s.transactionOperations.Transfer(s.ctx, transferParams)
		require.NoError(s.T(), err)

	})
}
