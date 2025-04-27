package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// User представляет пользователя
type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" validate:"required,alphanum,min=3,max=50" db:"username"`
	Email     string    `json:"email" validate:"required,email" db:"email"`
	Password  string    `json:"-" validate:"required,min=8" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Account представляет банковский счет
type Account struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Balance   float64   `json:"balance" db:"balance"`
	Currency  string    `json:"currency" db:"currency"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Card представляет банковскую карту
type Card struct {
	ID           int64     `json:"id" db:"id"`
	AccountID    int64     `json:"account_id" db:"account_id"`
	CardNumber   string    `json:"card_number" db:"card_number"` // Зашифровано
	ExpiryDate   string    `json:"expiry_date" db:"expiry_date"` // Зашифровано
	CVV          string    `json:"-" db:"cvv"`                   // Хешировано
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	HMAC         string    `json:"-" db:"hmac"` // Для проверки целостности
}

// Transaction представляет транзакцию
type Transaction struct {
	ID            int64     `json:"id" db:"id"`
	AccountID     int64     `json:"account_id" db:"account_id"`
	Amount        float64   `json:"amount" db:"amount"`
	Type          string    `json:"type" db:"type"` // deposit, withdrawal, transfer
	Description   string    `json:"description" db:"description"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Credit представляет кредит
type Credit struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	Amount       float64   `json:"amount" db:"amount"`
	InterestRate float64   `json:"interest_rate" db:"interest_rate"`
	TermMonths   int       `json:"term_months" db:"term_months"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// PaymentSchedule представляет график платежей по кредиту
type PaymentSchedule struct {
	ID           int64     `json:"id" db:"id"`
	CreditID     int64     `json:"credit_id" db:"credit_id"`
	PaymentDate  time.Time `json:"payment_date" db:"payment_date"`
	Amount       float64   `json:"amount" db:"amount"`
	Paid         bool      `json:"paid" db:"paid"`
	Penalty      float64   `json:"penalty" db:"penalty"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Validate валидирует структуру
func (u *User) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}