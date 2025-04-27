package services

import (
	"context"
	"errors"
	"time"

	"github.com/bank-service/internal/models"
	"github.com/bank-service/internal/repositories"
)

type AccountService interface {
	CreateAccount(ctx context.Context, userID int64) (*models.Account, error)
	GetAccounts(ctx context.Context, userID int64) ([]*models.Account, error)
}

type accountService struct {
	accountRepo repositories.AccountRepository
	userRepo    repositories.UserRepository
}

func NewAccountService(accountRepo repositories.AccountRepository, userRepo repositories.UserRepository) AccountService {
	return &accountService{
		accountRepo: accountRepo,
		userRepo:    userRepo,
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