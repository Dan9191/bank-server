package models

import (
	"errors"
	"regexp"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) Validate() error {
	if u.Username == "" || len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if u.Email == "" || !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(u.Email) {
		return errors.New("invalid email format")
	}
	if u.Password == "" || len(u.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID          int64     `json:"id"`
	AccountID   int64     `json:"account_id"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Card struct {
	ID         int64     `json:"id"`
	AccountID  int64     `json:"account_id"`
	CardNumber string    `json:"card_number"`
	ExpiryDate string    `json:"expiry_date"`
	CVV        string    `json:"cvv"`
	HMAC       string    `json:"hmac"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (c *Card) Validate() error {
	if c.AccountID <= 0 {
		return errors.New("invalid account ID")
	}
	if !regexp.MustCompile(`^\d{16}$`).MatchString(c.CardNumber) {
		return errors.New("card number must be 16 digits")
	}
	if !regexp.MustCompile(`^(0[1-9]|1[0-2])\/\d{2}$`).MatchString(c.ExpiryDate) {
		return errors.New("invalid expiry date format (MM/YY)")
	}
	if !regexp.MustCompile(`^\d{3}$`).MatchString(c.CVV) {
		return errors.New("CVV must be 3 digits")
	}
	return nil
}

type Credit struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Amount       float64   `json:"amount"`
	InterestRate float64   `json:"interest_rate"`
	TermMonths   int       `json:"term_months"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (c *Credit) Validate() error {
	if c.UserID <= 0 {
		return errors.New("invalid user ID")
	}
	if c.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	if c.InterestRate < 0 || c.InterestRate > 100 {
		return errors.New("interest rate must be between 0 and 100")
	}
	if c.TermMonths <= 0 {
		return errors.New("term months must be positive")
	}
	return nil
}

type PaymentSchedule struct {
	ID          int64     `json:"id"`
	CreditID    int64     `json:"credit_id"`
	PaymentDate time.Time `json:"payment_date"`
	Amount      float64   `json:"amount"`
	Paid        bool      `json:"paid"`
	Penalty     float64   `json:"penalty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ps *PaymentSchedule) Validate() error {
	if ps.CreditID <= 0 {
		return errors.New("invalid credit ID")
	}
	if ps.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	if ps.PaymentDate.IsZero() {
		return errors.New("payment date is required")
	}
	return nil
}
