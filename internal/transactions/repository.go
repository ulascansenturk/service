package transactions

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	"ulascansenturk/service/internal/constants"
)

type Repository interface {
	Create(ctx context.Context, transaction *Transaction) (*Transaction, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetByReferenceID(ctx context.Context, referenceID uuid.UUID) (*Transaction, error)
	GetByFromAccountID(ctx context.Context, fromAccountID uuid.UUID) ([]*Transaction, error)
	GetByToAccountID(ctx context.Context, toAccountID uuid.UUID) ([]*Transaction, error)
	GetByCreatedAt(ctx context.Context, createdAt time.Time) ([]*Transaction, error)
	Update(ctx context.Context, transaction *Transaction) error
	Transaction(ctx context.Context, fn func(*gorm.DB) (interface{}, error)) (interface{}, error)
	GetByIDForUpdate(ctx context.Context, transactionID uuid.UUID, tx *gorm.DB) (*Transaction, error)
	UpdateStatusWithTx(ctx context.Context, transaction Transaction, status constants.TransactionStatus, tx *gorm.DB) (*Transaction, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DB() *gorm.DB
}

// SQLRepository is the SQL implementation of TransactionRepository
type SQLRepository struct {
	db *gorm.DB
}

// NewSQLRepository creates a new SQLRepository
func NewSQLRepository(db *gorm.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) DB() *gorm.DB {
	return r.db
}
func (r *SQLRepository) Create(ctx context.Context, transaction *Transaction) (*Transaction, error) {
	if err := r.db.WithContext(ctx).Create(transaction).Error; err != nil {
		return nil, err
	}
	return transaction, nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	var transaction Transaction
	if err := r.db.WithContext(ctx).First(&transaction, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *SQLRepository) GetByReferenceID(ctx context.Context, referenceID uuid.UUID) (*Transaction, error) {
	var transaction Transaction
	if err := r.db.WithContext(ctx).Table("transactions").First(&transaction, "reference_id= ?", referenceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *SQLRepository) GetByFromAccountID(ctx context.Context, fromAccountID uuid.UUID) ([]*Transaction, error) {
	var transactions []*Transaction
	if err := r.db.WithContext(ctx).Where("from_account_id = ?", fromAccountID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *SQLRepository) GetByToAccountID(ctx context.Context, toAccountID uuid.UUID) ([]*Transaction, error) {
	var transactions []*Transaction
	if err := r.db.WithContext(ctx).Where("to_account_id = ?", toAccountID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *SQLRepository) GetByCreatedAt(ctx context.Context, createdAt time.Time) ([]*Transaction, error) {
	var transactions []*Transaction
	if err := r.db.WithContext(ctx).Where("created_at = ?", createdAt).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *SQLRepository) Update(ctx context.Context, transaction *Transaction) error {
	if err := r.db.WithContext(ctx).Save(transaction).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&Transaction{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLRepository) UpdateStatusWithTx(ctx context.Context, transaction Transaction, status constants.TransactionStatus, tx *gorm.DB) (*Transaction, error) {
	if tx == nil {
		return nil, errors.New("transaction is required")
	}

	// Update the status
	if err := tx.WithContext(ctx).Model(&transaction).Update("status", status).Error; err != nil {
		return nil, err
	}

	// Fetch the updated transaction
	var updatedTransaction Transaction
	if err := tx.WithContext(ctx).First(&updatedTransaction, transaction.ID).Error; err != nil {
		return nil, err
	}

	return &updatedTransaction, nil
}
func (r *SQLRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = fn(tx)
		return err
	})
	return result, err
}
func (r *SQLRepository) GetByIDForUpdate(ctx context.Context, transactionID uuid.UUID, tx *gorm.DB) (*Transaction, error) {
	var transaction Transaction

	if tx == nil {
		return nil, errors.New("transaction is required")
	}

	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&transaction, "id = ?", transactionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &transaction, nil
}
