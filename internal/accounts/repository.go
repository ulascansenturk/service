package accounts

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, account *Account) (*Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateWithTx(ctx context.Context, account *Account, tx *gorm.DB) error

	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Account, error)
}

type SQLRepository struct {
	db *gorm.DB
}

// NewSQLRepository creates a new SQLRepository
func NewSQLRepository(db *gorm.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, account *Account) (*Account, error) {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	var account Account
	if err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *SQLRepository) Update(ctx context.Context, account *Account) error {
	if err := r.db.WithContext(ctx).Save(account).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&Account{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Account, error) {
	var accounts []*Account
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}
func (r *SQLRepository) UpdateWithTx(ctx context.Context, account *Account, tx *gorm.DB) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	if err := tx.WithContext(ctx).Model(&Account{}).Where("id = ?", account.ID).Updates(account).Error; err != nil {
		return err
	}
	return nil
}
