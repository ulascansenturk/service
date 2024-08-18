package transactions

import (
	"github.com/google/uuid"
	"time"

	"gorm.io/datatypes"
	"ulascansenturk/service/internal/constants"
)

type Transaction struct {
	ID              uuid.UUID                   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id,omitempty"`
	UserID          *uuid.UUID                  `gorm:"type:uuid" json:"user_id,omitempty"`
	Amount          int                         `gorm:"type:integer" json:"amount"`
	AccountID       uuid.UUID                   `gorm:"type:uuid" json:"account_id"`
	CurrencyCode    constants.CurrencyCode      `gorm:"type:varchar(3)" json:"currency_code"`
	ReferenceID     uuid.UUID                   `gorm:"type:uuid" json:"reference_id"`
	Metadata        datatypes.JSONMap           `gorm:"type:jsonb" json:"metadata"`
	Status          constants.TransactionStatus `gorm:"type:varchar(50)" json:"status"`
	TransactionType constants.TransactionType   `gorm:"type:varchar(50)" json:"transaction_type"`
	CreatedAt       time.Time                   `gorm:"type:timestamptz;default:now()" json:"created_at,omitempty"`
	UpdatedAt       time.Time                   `gorm:"type:timestamptz;default:now();autoUpdateTime()" json:"updated_at,omitempty"`
}
