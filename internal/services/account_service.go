package services

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/bank-service/internal/models"
	"github.com/bank-service/internal/repositories"
)

type accountService struct {
	accountRepo     repositories.AccountRepository
	userRepo        repositories.UserRepository
	transactionRepo repositories.TransactionRepository
	db              *sql.DB
	mutex           sync.Mutex
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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	account := &models.Account{
		UserID:    userID,
		Balance:   0.0,
		Currency:  "RUB",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.accountRepo.Create(ctx, account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *accountService) GetAccounts(ctx context.Context, userID int64) ([]*models.Account, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	accounts, err := s.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *accountService) Deposit(ctx context.Context, accountID int64, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("account not found")
	}

	newBalance := account.Balance + amount
	err = s.accountRepo.UpdateBalance(ctx, accountID, newBalance)
	if err != nil {
		return err
	}

	transaction := &models.Transaction{
		AccountID:   accountID,
		Amount:      amount,
		Type:        "deposit",
		Description: "Deposit",
		CreatedAt:   time.Now(),
	}
	err = s.transactionRepo.Create(ctx, tx, transaction)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *accountService) Withdraw(ctx context.Context, accountID int64, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("account not found")
	}

	if account.Balance < amount {
		return errors.New("insufficient funds")
	}

	newBalance := account.Balance - amount
	err = s.accountRepo.UpdateBalance(ctx, accountID, newBalance)
	if err != nil {
		return err
	}

	transaction := &models.Transaction{
		AccountID:   accountID,
		Amount:      -amount,
		Type:        "withdrawal",
		Description: "Withdrawal",
		CreatedAt:   time.Now(),
	}
	err = s.transactionRepo.Create(ctx, tx, transaction)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *accountService) Transfer(ctx context.Context, fromAccountID, toAccountID int64, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	fromAccount, err := s.accountRepo.FindByID(ctx, fromAccountID)
	if err != nil {
		return err
	}
	if fromAccount == nil {
		return errors.New("source account not found")
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

	err = s.accountRepo.UpdateBalance(ctx, fromAccountID, fromAccount.Balance-amount)
	if err != nil {
		return err
	}

	err = s.accountRepo.UpdateBalance(ctx, toAccountID, toAccount.Balance+amount)
	if err != nil {
		return err
	}

	fromTransaction := &models.Transaction{
		AccountID:   fromAccountID,
		Amount:      -amount,
		Type:        "transfer_out",
		Description: "Transfer to account " + string(toAccountID),
		CreatedAt:   time.Now(),
	}
	err = s.transactionRepo.Create(ctx, tx, fromTransaction)
	if err != nil {
		return err
	}

	toTransaction := &models.Transaction{
		AccountID:   toAccountID,
		Amount:      amount,
		Type:        "transfer_in",
		Description: "Transfer from account " + string(fromAccountID),
		CreatedAt:   time.Now(),
	}
	err = s.transactionRepo.Create(ctx, tx, toTransaction)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *accountService) GetTransactions(ctx context.Context, accountID, userID int64) ([]*models.Transaction, error) {
	// Проверяем, существует ли счёт и принадлежит ли он пользователю
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, errors.New("account not found")
	}
	if account.UserID != userID {
		return nil, errors.New("unauthorized access to account")
	}

	// Получаем транзакции
	transactions, err := s.transactionRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}
