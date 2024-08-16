package users

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SQLRepository struct {
	db *gorm.DB
}

func NewSQLRepository(db *gorm.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, user *User) (*User, error) {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No record found
		}
		return nil, err
	}
	return &user, nil
}

func (r *SQLRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *SQLRepository) Update(ctx context.Context, user *User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&User{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
