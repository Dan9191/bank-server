package repositories

import (
	"context"
	"database/sql"

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

// TransactionRepository определяет методы для работы с транзакциями
type TransactionRepository interface {
	Create(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error
	FindByAccountID(ctx context.Context, accountID int64) ([]*models.Transaction, error)
}

// CardRepository определяет методы для работы с картами
type CardRepository interface {
	Create(ctx context.Context, card *models.Card) error
	FindByAccountID(ctx context.Context, accountID int64) ([]*models.Card, error)
}

// CreditRepository определяет методы для работы с кредитами и графиком платежей
type CreditRepository interface {
	CreateCredit(ctx context.Context, credit *models.Credit) error
	FindByUserID(ctx context.Context, userID int64) ([]*models.Credit, error)
	CreatePaymentSchedule(ctx context.Context, paymentSchedule *models.PaymentSchedule) error
	FindPaymentSchedulesByCreditID(ctx context.Context, creditID int64) ([]*models.PaymentSchedule, error)
}
