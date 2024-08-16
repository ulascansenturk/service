package users

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	CreateUser(ctx context.Context, user *User, password string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type UserServiceImpl struct {
	repo Repository
}

func NewUserService(repo Repository) *UserServiceImpl {
	return &UserServiceImpl{repo: repo}
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, user *User, password string) (*User, error) {
	if user.Email == "" {
		return nil, errors.New("email is required")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = hashedPassword

	return s.repo.Create(ctx, user)
}

func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *UserServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserServiceImpl) UpdateUser(ctx context.Context, user *User) error {
	// Add any business logic or validation here if needed
	if user.ID == uuid.Nil {
		return errors.New("invalid user ID")
	}
	return s.repo.Update(ctx, user)
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// You could check if the user exists before deleting, if needed
	return s.repo.Delete(ctx, id)
}
