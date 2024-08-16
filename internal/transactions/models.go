package transactions

import (
	"github.com/google/uuid"
	"time"
	"ulascansenturk/service/internal/constants"
)

type Transaction struct {
	ID              uuid.UUID                   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id,omitempty"`
	UserID          *uuid.UUID                  `gorm:"type:uuid" json:"user_id,omitempty"`
	Amount          int                         `gorm:"type:integer" json:"amount"`
	AccountID       uuid.UUID                   `gorm:"type:uuid" json:"account_id"`
	CurrencyCode    constants.CurrencyCode      `gorm:"type:varchar(3)" json:"currency_code"`
	ReferenceID     uuid.UUID                   `gorm:"type:uuid" json:"reference_id"`
	Metadata        *map[string]interface{}     `gorm:"type:jsonb" json:"metadata"`
	Status          constants.TransactionStatus `gorm:"type:varchar(50)" json:"status"`
	TransactionType constants.TransactionType   `gorm:"type:varchar(50)" json:"transaction_type"`
	CreatedAt       time.Time                   `gorm:"type:timestamptz;default:now()" json:"created_at,omitempty"`
	UpdatedAt       time.Time                   `gorm:"type:timestamptz;default:now();autoUpdateTime()" json:"updated_at,omitempty"`
}
type DBTransaction struct {
	ID              uuid.UUID                   `json:"id,omitempty"`
	UserID          *uuid.UUID                  `json:"user_id,omitempty"`
	Amount          int                         `json:"amount"`
	AccountID       uuid.UUID                   `json:"account_id"`
	CurrencyCode    constants.CurrencyCode      `json:"currency_code"`
	ReferenceID     uuid.UUID                   `json:"reference_id"`
	Metadata        *map[string]interface{}     `json:"metadata"`
	Status          constants.TransactionStatus `json:"status"`
	TransactionType constants.TransactionType   `json:"transaction_type"`
	CreatedAt       time.Time                   `json:"created_at,omitempty"`
	UpdatedAt       time.Time                   `json:"updated_at,omitempty"`
}

func FromDBTransaction(dbTransaction *DBTransaction) *Transaction {
	return &Transaction{
		ID:              dbTransaction.ID,
		UserID:          dbTransaction.UserID,
		Amount:          dbTransaction.Amount,
		CurrencyCode:    dbTransaction.CurrencyCode,
		AccountID:       dbTransaction.AccountID,
		ReferenceID:     dbTransaction.ReferenceID,
		Metadata:        dbTransaction.Metadata,
		Status:          dbTransaction.Status,
		TransactionType: dbTransaction.TransactionType,
		CreatedAt:       dbTransaction.CreatedAt,
		UpdatedAt:       dbTransaction.UpdatedAt,
	}
}
