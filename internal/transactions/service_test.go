package transactions_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ulascansenturk/service/internal/constants"
	"ulascansenturk/service/internal/transactions"
)

func TestTransactionService_GetTransactionByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := transactions.NewSQLRepository(gormDB)
	service := transactions.NewTransactionService(repo, validator.New())

	ctx := context.Background()
	transactionID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "amount", "status"}).
		AddRow(transactionID, 100.0, constants.TransactionStatusPENDING)

	mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE id = \$1`).
		WithArgs(transactionID).
		WillReturnRows(rows)

	transaction, err := service.GetTransactionByID(ctx, transactionID)
	require.NoError(t, err)
	assert.Equal(t, transactionID, transaction.ID)
	assert.Equal(t, constants.TransactionStatusPENDING, transaction.Status)
	assert.Equal(t, 100.0, transaction.Amount)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionService_GetTransactionByReferenceID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := transactions.NewSQLRepository(gormDB)
	service := transactions.NewTransactionService(repo, validator.New())

	ctx := context.Background()
	referenceID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "reference_id", "amount", "status"}).
		AddRow(uuid.New(), referenceID, 200.0, constants.TransactionStatusSUCCESS)

	mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE reference_id = \$1`).
		WithArgs(referenceID).
		WillReturnRows(rows)

	transaction, err := service.GetTransactionByReferenceID(ctx, referenceID)
	require.NoError(t, err)
	assert.Equal(t, referenceID, transaction.ReferenceID)
	assert.Equal(t, constants.TransactionStatusSUCCESS, transaction.Status)
	assert.Equal(t, 200.0, transaction.Amount)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionService_UpdateTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := transactions.NewSQLRepository(gormDB)
	service := transactions.NewTransactionService(repo, validator.New())

	ctx := context.Background()
	transaction := &transactions.Transaction{
		ID:     uuid.New(),
		Amount: 300.0,
		Status: constants.TransactionStatusPENDING,
	}

	mock.ExpectExec(`UPDATE "transactions" SET "amount"=\$1,"status"=\$2 WHERE "id"=\$3`).
		WithArgs(transaction.Amount, transaction.Status, transaction.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = service.UpdateTransaction(ctx, transaction, nil)
	require.NoError(t, err)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionService_UpdateTransactionStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := transactions.NewSQLRepository(gormDB)
	service := transactions.NewTransactionService(repo, validator.New())

	ctx := context.Background()
	transactionID := uuid.New()
	newStatus := constants.TransactionStatusSUCCESS

	// Mock getting the transaction
	rows := sqlmock.NewRows([]string{"id", "status"}).
		AddRow(transactionID, constants.TransactionStatusPENDING)

	mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE id = \$1 FOR UPDATE`).
		WithArgs(transactionID).
		WillReturnRows(rows)

	// Mock updating the transaction status
	mock.ExpectExec(`UPDATE "transactions" SET "status" = \$1 WHERE "id" = \$2`).
		WithArgs(newStatus, transactionID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	updatedTransaction, err := service.UpdateTransactionStatus(ctx, transactionID, newStatus)
	require.NoError(t, err)
	assert.Equal(t, newStatus, updatedTransaction.Status)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionService_BeginTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := transactions.NewSQLRepository(gormDB)
	service := transactions.NewTransactionService(repo, validator.New())

	ctx := context.Background()

	mock.ExpectBegin()

	tx, err := service.BeginTransaction(ctx)
	require.NoError(t, err)
	assert.NotNil(t, tx)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}
