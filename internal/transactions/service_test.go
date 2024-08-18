package transactions_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"

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

	// Fixed the query to account for LIMIT clause
	mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE id = \$1 ORDER BY "transactions"."id" LIMIT \$2`).
		WithArgs(transactionID, 1).
		WillReturnRows(rows)

	transaction, err := service.GetTransactionByID(ctx, transactionID)
	require.NoError(t, err)
	assert.Equal(t, transactionID, transaction.ID)
	assert.Equal(t, constants.TransactionStatusPENDING, transaction.Status)
	assert.Equal(t, 100, transaction.Amount)

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
		AddRow(uuid.New(), referenceID, 200, constants.TransactionStatusSUCCESS)

	// Adjusted the query regex to account for the space after '='
	mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE reference_id= \$1 ORDER BY "transactions"."id" LIMIT \$2`).
		WithArgs(referenceID, 1).
		WillReturnRows(rows)

	transaction, err := service.GetTransactionByReferenceID(ctx, referenceID)
	require.NoError(t, err)
	assert.Equal(t, referenceID, transaction.ReferenceID)
	assert.Equal(t, constants.TransactionStatusSUCCESS, transaction.Status)
	assert.Equal(t, 200, transaction.Amount)

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

	// Ensure transaction start is mocked
	mock.ExpectBegin()

	tx, err := service.BeginTransaction(ctx)
	require.NoError(t, err)
	assert.NotNil(t, tx)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}
