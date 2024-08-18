package transactions

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
	"ulascansenturk/service/internal/constants"
)

type Service interface {
	GetTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetTransactionByReferenceID(ctx context.Context, referenceID uuid.UUID) (*Transaction, error)
	GetTransactionsByFromAccountID(ctx context.Context, fromAccountID uuid.UUID) ([]*Transaction, error)
	GetTransactionsByToAccountID(ctx context.Context, toAccountID uuid.UUID) ([]*Transaction, error)
	GetTransactionsByCreatedAt(ctx context.Context, createdAt time.Time) ([]*Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *Transaction, tx *gorm.DB) error
	DeleteTransaction(ctx context.Context, id uuid.UUID) error
	UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status constants.TransactionStatus) (*Transaction, error)
	BeginTransaction(ctx context.Context) (*gorm.DB, error) // New method
}

type TransactionServiceImpl struct {
	repo     Repository
	validate *validator.Validate
}

func NewTransactionService(repo Repository, validate *validator.Validate) *TransactionServiceImpl {
	return &TransactionServiceImpl{repo: repo, validate: validate}
}

func (s *TransactionServiceImpl) GetTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if transaction == nil {
		return nil, errors.New("transaction not found")
	}
	return transaction, nil
}

func (s *TransactionServiceImpl) GetTransactionByReferenceID(ctx context.Context, referenceID uuid.UUID) (*Transaction, error) {
	transaction, err := s.repo.GetByReferenceID(ctx, referenceID)
	if err != nil {
		return nil, err
	}
	if transaction == nil {
		return nil, errors.New("transaction not found")
	}
	return transaction, nil
}

func (s *TransactionServiceImpl) GetTransactionsByFromAccountID(ctx context.Context, fromAccountID uuid.UUID) ([]*Transaction, error) {
	transactions, err := s.repo.GetByFromAccountID(ctx, fromAccountID)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (s *TransactionServiceImpl) GetTransactionsByToAccountID(ctx context.Context, toAccountID uuid.UUID) ([]*Transaction, error) {
	transactions, err := s.repo.GetByToAccountID(ctx, toAccountID)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (s *TransactionServiceImpl) GetTransactionsByCreatedAt(ctx context.Context, createdAt time.Time) ([]*Transaction, error) {
	transactions, err := s.repo.GetByCreatedAt(ctx, createdAt)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (s *TransactionServiceImpl) UpdateTransaction(ctx context.Context, transaction *Transaction, tx *gorm.DB) error {
	if transaction.ID == uuid.Nil {
		return errors.New("invalid transaction ID")
	}
	if transaction.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	if tx == nil {
		return s.repo.Update(ctx, transaction)
	}

	// Call the repository to update the transaction
	return s.repo.Update(ctx, transaction)
}

func (s *TransactionServiceImpl) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	// You might want to check if the transaction exists before deleting
	return s.repo.Delete(ctx, id)
}

// BeginTransaction starts a new database transaction
func (s *TransactionServiceImpl) BeginTransaction(ctx context.Context) (*gorm.DB, error) {
	tx := s.repo.DB().Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}
func (s *TransactionServiceImpl) UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status constants.TransactionStatus) (*Transaction, error) {
	result, err := s.repo.Transaction(ctx, func(tx *gorm.DB) (interface{}, error) {
		transaction, err := s.repo.GetByIDForUpdate(ctx, id, tx)
		if err != nil {
			return nil, err
		}
		if transaction == nil {
			return nil, errors.New("transaction not found")
		}

		if status == "" {
			return nil, errors.New("status cannot be empty")
		}

		updatedTrx, updateTrxErr := s.repo.UpdateStatusWithTx(ctx, *transaction, status, tx)
		if updateTrxErr != nil {
			return nil, updateTrxErr
		}

		return updatedTrx, nil
	})

	if err != nil {
		return nil, err
	}

	updatedTransaction, ok := result.(*Transaction)
	if !ok {
		return nil, errors.New("unexpected result type")
	}

	return updatedTransaction, nil
}
