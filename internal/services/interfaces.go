package services

import (
	"context"

	"github.com/bank-service/internal/models"
)

// UserService определяет методы для работы с пользователями
type UserService interface {
	Register(ctx context.Context, username, email, password string) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetProfile(ctx context.Context, userID int64) (*models.User, error)
}

// AccountService определяет методы для работы со счетами
type AccountService interface {
	CreateAccount(ctx context.Context, userID int64) (*models.Account, error)
	GetAccounts(ctx context.Context, userID int64) ([]*models.Account, error)
	Deposit(ctx context.Context, accountID int64, amount float64) error
	Withdraw(ctx context.Context, accountID int64, amount float64) error
	Transfer(ctx context.Context, fromAccountID, toAccountID int64, amount float64) error
	GetTransactions(ctx context.Context, accountID, userID int64) ([]*models.Transaction, error)
}

// CardService определяет методы для работы с картами
type CardService interface {
	CreateCard(ctx context.Context, accountID int64, cardNumber, expiryDate, cvv string) (*models.Card, error)
	GetCards(ctx context.Context, accountID int64) ([]*models.Card, error)
}

// CreditService определяет методы для работы с кредитами
type CreditService interface {
	CreateCredit(ctx context.Context, userID int64, amount, interestRate float64, termMonths int) (*models.Credit, error)
	GetCredits(ctx context.Context, userID int64) ([]*models.Credit, error)
	GetPaymentSchedules(ctx context.Context, creditID, userID int64) ([]*models.PaymentSchedule, error)
}
