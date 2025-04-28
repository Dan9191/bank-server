package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/bank-service/internal/models"
	"github.com/bank-service/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type cardService struct {
	cardRepo    repositories.CardRepository
	accountRepo repositories.AccountRepository
	hmacSecret  string
}

func NewCardService(cardRepo repositories.CardRepository, accountRepo repositories.AccountRepository, hmacSecret string) CardService {
	return &cardService{
		cardRepo:    cardRepo,
		accountRepo: accountRepo,
		hmacSecret:  hmacSecret,
	}
}

func (s *cardService) CreateCard(ctx context.Context, accountID int64, cardNumber, expiryDate, cvv string) (*models.Card, error) {
	// Проверяем существование счёта
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, errors.New("account not found")
	}

	// Создаём карту
	card := &models.Card{
		AccountID:  accountID,
		CardNumber: cardNumber,
		ExpiryDate: expiryDate,
		CVV:        cvv,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Валидируем карту
	if err := card.Validate(); err != nil {
		return nil, err
	}

	// Вычисляем HMAC для card_number + expiry_date
	hmacData := card.CardNumber + card.ExpiryDate
	mac := hmac.New(sha256.New, []byte(s.hmacSecret))
	mac.Write([]byte(hmacData))
	card.HMAC = hex.EncodeToString(mac.Sum(nil))

	// Хешируем CVV
	hashedCVV, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	card.CVV = string(hashedCVV)

	// Сохраняем карту
	if err := s.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}

	return card, nil
}

func (s *cardService) GetCards(ctx context.Context, accountID int64) ([]*models.Card, error) {
	// Проверяем существование счёта
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, errors.New("account not found")
	}

	// Получаем карты
	cards, err := s.cardRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return cards, nil
}
