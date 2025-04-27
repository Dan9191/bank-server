package repositories

import (
	"context"

	"github.com/bank-service/internal/models"
)

// UserRepository определяет методы для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id int64) (*models.User, error)
}

// AccountRepository определяет методы для работы со счетами
type AccountRepository interface {
	Create(ctx context.Context, account *models.Account) error
	FindByID(ctx context.Context, id int64) (*models.Account, error)
	FindByUserID(ctx context.Context, userID int64) ([]*models.Account, error)
	UpdateBalance(ctx context.Context, accountID int64, balance float64) error
}