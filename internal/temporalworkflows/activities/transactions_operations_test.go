package activities

import (
	"context"
	"gorm.io/datatypes"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"ulascansenturk/service/internal/accounts"
	accountMocks "ulascansenturk/service/internal/accounts/mocks"
	"ulascansenturk/service/internal/constants"
	mockTime "ulascansenturk/service/internal/helpers/mocks"
	"ulascansenturk/service/internal/transactions"
	"ulascansenturk/service/internal/transactions/mocks"
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
	s.Run("Process Transfer Activity", func() {
		sourceAccID := uuid.New()
		destinationAccID := uuid.New()
		sourceAccountUser := users.User{
			ID:        uuid.New(),
			Email:     "ulas@gmail.com",
			FirstName: "ulas",
			IsActive:  true,
		}

		sourceAccount := accounts.Account{
			ID:        sourceAccID,
			UserID:    sourceAccountUser.ID,
			Balance:   1000,
			Currency:  "USD",
			Status:    constants.AccountStatusACTIVE,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		destinationAccountUser := users.User{
			ID:        uuid.New(),
			Email:     "meric@gmail.com",
			FirstName: "meric",
			IsActive:  true,
		}

		destinationAccount := accounts.Account{
			ID:        destinationAccID,
			UserID:    destinationAccountUser.ID,
			Balance:   500,
			Currency:  "USD",
			Status:    constants.AccountStatusACTIVE,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		amount := 100
		feeAmount := 10

		params := TransferParams{
			Amount:                            amount,
			FeeAmount:                         &feeAmount,
			DestinationAccountID:              destinationAccID,
			SourceTransactionReferenceID:      uuid.New(),
			DestinationTransactionReferenceID: uuid.New(),
			FeeTransactionReferenceID:         uuid.New(),
			SourceAccountID:                   sourceAccID,
		}

		timestamp := time.Now()
		s.timeProvider.On("Now").Return(timestamp)

		s.accountsService.On("GetAccountByID", mock.Anything, sourceAccID).Return(&sourceAccount, nil)
		s.accountsService.On("GetAccountByID", mock.Anything, destinationAccID).Return(&destinationAccount, nil)

		s.accountsService.On("UpdateBalance", mock.Anything, sourceAccID, amount, constants.BalanceOperationDECREASE.String()).Return(nil)
		s.accountsService.On("UpdateBalance", mock.Anything, destinationAccID, amount, constants.BalanceOperationINCREASE.String()).Return(nil)

		sourceTransaction := &transactions.Transaction{
			ID:              uuid.New(),
			AccountID:       sourceAccID,
			Amount:          amount,
			Status:          constants.TransactionStatusPENDING,
			TransactionType: constants.TransactionTypeOUTBOUND,
			Metadata: datatypes.JSONMap(map[string]interface{}{
				"OperationType":        "Transfer",
				"LinkedTransactionID":  params.SourceTransactionReferenceID.String(),
				"LinkedAccountID":      sourceAccID.String(),
				"DestinationAccountID": destinationAccID.String(),
				"timestamp":            timestamp.Format(time.RFC3339),
			}),
		}

		destinationTransaction := &transactions.Transaction{
			ID:              uuid.New(),
			AccountID:       destinationAccID,
			Amount:          amount,
			Status:          constants.TransactionStatusPENDING,
			TransactionType: constants.TransactionTypeINBOUND,
			Metadata: datatypes.JSONMap(map[string]interface{}{
				"OperationType":       "Transfer",
				"LinkedTransactionID": params.DestinationTransactionReferenceID.String(),
				"LinkedAccountID":     destinationAccID.String(),
				"SourceAccountID":     sourceAccID.String(),
				"timestamp":           timestamp.Format(time.RFC3339),
			}),
		}

		feeTransaction := &transactions.Transaction{
			ID:              uuid.New(),
			AccountID:       sourceAccID,
			Amount:          feeAmount,
			Status:          constants.TransactionStatusPENDING,
			TransactionType: constants.TransactionTypeOUTGOINGFEE,
			Metadata: datatypes.JSONMap(map[string]interface{}{
				"OperationType":       "Fee Transfer",
				"LinkedTransactionID": params.FeeTransactionReferenceID.String(),
				"LinkedAccountID":     sourceAccID.String(),
				"timestamp":           timestamp.Format(time.RFC3339),
			}),
		}

		s.finderOrCreatorService.On("Call", mock.Anything, mock.Anything).Return(sourceTransaction, nil).Once()
		s.finderOrCreatorService.On("Call", mock.Anything, mock.Anything).Return(destinationTransaction, nil).Once()
		s.finderOrCreatorService.On("Call", mock.Anything, mock.Anything).Return(feeTransaction, nil).Once()

		s.transactionsService.On("UpdateTransactionStatus", mock.Anything, sourceTransaction.ID, constants.TransactionStatusSUCCESS).Return(sourceTransaction, nil)
		s.transactionsService.On("UpdateTransactionStatus", mock.Anything, destinationTransaction.ID, constants.TransactionStatusSUCCESS).Return(destinationTransaction, nil)
		s.transactionsService.On("UpdateTransactionStatus", mock.Anything, feeTransaction.ID, constants.TransactionStatusSUCCESS).Return(feeTransaction, nil)

		result, err := s.transactionOperations.Transfer(s.ctx, params)
		require.NoError(s.T(), err)

		require.Equal(s.T(), params.SourceTransactionReferenceID, result.SourceTransactionReferenceID)
		require.Equal(s.T(), params.DestinationTransactionReferenceID, result.DestinationTransactionReferenceID)
		require.Equal(s.T(), params.FeeTransactionReferenceID, result.FeeTransactionReferenceID)
		require.NotNil(s.T(), result.SourceTransaction)
		require.NotNil(s.T(), result.DestinationTransaction)
		require.NotNil(s.T(), result.FeeTransaction)
	})
}
