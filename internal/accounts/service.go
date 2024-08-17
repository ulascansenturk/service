package accounts

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	CreateAccount(ctx context.Context, account *Account) (*Account, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error)
	UpdateAccount(ctx context.Context, account *Account, tx *gorm.DB) error
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]*Account, error)
	UpdateBalance(ctx context.Context, accountID uuid.UUID, amount int, operation string) error
}

type AccountServiceImpl struct {
	repo     Repository
	validate *validator.Validate
}

func NewUserBankAccountService(repo Repository, validate *validator.Validate) *AccountServiceImpl {
	return &AccountServiceImpl{repo: repo, validate: validate}
}

func (s *AccountServiceImpl) CreateAccount(ctx context.Context, account *Account) (*Account, error) {
	if err := s.validate.Struct(account); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, account)
}

func (s *AccountServiceImpl) GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, errors.New("account not found")
	}
	return account, nil
}

// UpdateAccount updates an account, optionally using a transaction
func (s *AccountServiceImpl) UpdateAccount(ctx context.Context, account *Account, tx *gorm.DB) error {
	if account.ID == uuid.Nil {
		return errors.New("invalid account ID")
	}

	if account.Balance < 0 {
		return errors.New("balance cannot be negative")
	}

	if tx == nil {
		return s.repo.Update(ctx, account)
	}

	return s.repo.UpdateWithTx(ctx, account, tx)
}

func (s *AccountServiceImpl) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *AccountServiceImpl) GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]*Account, error) {
	accounts, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *AccountServiceImpl) UpdateBalance(ctx context.Context, accountID uuid.UUID, amount int, operation string) error {
	return s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		account, err := s.repo.GetByIDForUpdate(ctx, accountID, tx)
		if err != nil {
			return err
		}

		switch operation {
		case "INCREASE":
			account.Balance += amount
		case "DECREASE":
			if account.Balance < amount {
				return errors.New("insufficient funds")
			}
			account.Balance -= amount
		default:
			return errors.New("invalid operation")
		}

		return s.repo.UpdateWithTx(ctx, account, tx)
	})
}
