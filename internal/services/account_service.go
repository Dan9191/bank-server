package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bank-service/internal/models"
	"github.com/bank-service/internal/repositories"
)

type AccountService interface {
	CreateAccount(ctx context.Context, userID int64) (*models.Account, error)
	GetAccounts(ctx context.Context, userID int64) ([]*models.Account, error)
	Deposit(ctx context.Context, accountID int64, amount float64, userID int64) error
	Withdraw(ctx context.Context, accountID int64, amount float64, userID int64) error
	Transfer(ctx context.Context, fromAccountID, toAccountID int64, amount float64, userID int64) error
}

type accountService struct {
	accountRepo     repositories.AccountRepository
	userRepo        repositories.UserRepository
	transactionRepo repositories.TransactionRepository
	db              *sql.DB
}

func NewAccountService(accountRepo repositories.AccountRepository, userRepo repositories.UserRepository, transactionRepo repositories.TransactionRepository, db *sql.DB) AccountService {
	return &accountService{
		accountRepo:     accountRepo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		db:              db,
	}
}

func (s *accountService) CreateAccount(ctx context.Context, userID int64) (*models.Account, error) {
	// Проверяем, существует ли пользователь
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Создаем новый счет
	account := &models.Account{
		UserID:    userID,
		Balance:   0.0,
		Currency:  "RUB",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохраняем счет в базе
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *accountService) GetAccounts(ctx context.Context, userID int64) ([]*models.Account, error) {
	// Получаем все счета пользователя
	accounts, err := s.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *accountService) Deposit(ctx context.Context, accountID int64, amount float64, userID int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	// Проверяем, существует ли счет и принадлежит ли он пользователю
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("account not found")
	}
	if account.UserID != userID {
		return errors.New("unauthorized access to account")
	}

	// Начинаем транзакцию
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Обновляем баланс
	newBalance := account.Balance + amount
	if err := s.accountRepo.UpdateBalance(ctx, accountID, newBalance); err != nil {
		return err
	}

	// Записываем транзакцию
	transaction := &models.Transaction{
		AccountID:   accountID,
		Amount:      amount,
		Type:        "deposit",
		Description: "Deposit",
		CreatedAt:   time.Now(),
	}
	if err := s.transactionRepo.Create(ctx, tx, transaction); err != nil {
		return err
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *accountService) Withdraw(ctx context.Context, accountID int64, amount float64, userID int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	// Проверяем, существует ли счет и принадлежит ли он пользователю
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("account not found")
	}
	if account.UserID != userID {
		return errors.New("unauthorized access to account")
	}
	if account.Balance < amount {
		return errors.New("insufficient funds")
	}

	// Начинаем транзакцию
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Обновляем баланс
	newBalance := account.Balance - amount
	if err := s.accountRepo.UpdateBalance(ctx, accountID, newBalance); err != nil {
		return err
	}

	// Записываем транзакцию
	transaction := &models.Transaction{
		AccountID:   accountID,
		Amount:      -amount,
		Type:        "withdrawal",
		Description: "Withdrawal",
		CreatedAt:   time.Now(),
	}
	if err := s.transactionRepo.Create(ctx, tx, transaction); err != nil {
		return err
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *accountService) Transfer(ctx context.Context, fromAccountID, toAccountID int64, amount float64, userID int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	if fromAccountID == toAccountID {
		return errors.New("cannot transfer to the same account")
	}

	// Проверяем, существуют ли счета
	fromAccount, err := s.accountRepo.FindByID(ctx, fromAccountID)
	if err != nil {
		return err
	}
	if fromAccount == nil {
		return errors.New("source account not found")
	}
	if fromAccount.UserID != userID {
		return errors.New("unauthorized access to source account")
	}
	if fromAccount.Balance < amount {
		return errors.New("insufficient funds")
	}

	toAccount, err := s.accountRepo.FindByID(ctx, toAccountID)
	if err != nil {
		return err
	}
	if toAccount == nil {
		return errors.New("destination account not found")
	}

	// Начинаем транзакцию
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Обновляем баланс отправителя
	if err := s.accountRepo.UpdateBalance(ctx, fromAccountID, fromAccount.Balance-amount); err != nil {
		return err
	}

	// Обновляем баланс получателя
	if err := s.accountRepo.UpdateBalance(ctx, toAccountID, toAccount.Balance+amount); err != nil {
		return err
	}

	// Записываем транзакцию для отправителя
	fromTransaction := &models.Transaction{
		AccountID:   fromAccountID,
		Amount:      -amount,
		Type:        "transfer_out",
		Description: "Transfer to account " + string(toAccountID),
		CreatedAt:   time.Now(),
	}
	if err := s.transactionRepo.Create(ctx, tx, fromTransaction); err != nil {
		return err
	}

	// Записываем транзакцию для получателя
	toTransaction := &models.Transaction{
		AccountID:   toAccountID,
		Amount:      amount,
		Type:        "transfer_in",
		Description: "Transfer from account " + string(fromAccountID),
		CreatedAt:   time.Now(),
	}
	if err := s.transactionRepo.Create(ctx, tx, toTransaction); err != nil {
		return err
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
