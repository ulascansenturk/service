package accounts

import (
	"github.com/google/uuid"
	"time"
	"ulascansenturk/service/internal/constants"
)

type Account struct {
	ID        uuid.UUID               `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID               `gorm:"type:uuid;not null;index" validate:"required"`
	Balance   int                     `gorm:"not null;"`
	Currency  string                  `gorm:"type:varchar(3);not null" validate:"required,len=3,iso4217"`
	Status    constants.AccountStatus `gorm:"type:varchar(50);not null"`
	CreatedAt time.Time               `gorm:"type:timestamp with time zone;not null" `
	UpdatedAt time.Time               `gorm:"type:timestamp with time zone;not null" `
}
