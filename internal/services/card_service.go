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

type CardService interface {
	CreateCard(ctx context.Context, accountID int64, cardNumber, expiryDate, cvv string, userID int64) (*models.Card, error)
	GetCards(ctx context.Context, accountID int64, userID int64) ([]*models.Card, error)
}

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

func (s *cardService) CreateCard(ctx context.Context, accountID int64, cardNumber, expiryDate, cvv string, userID int64) (*models.Card, error) {
	// Проверяем, существует ли счет и принадлежит ли он пользователю
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

	// Создаем карту
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

	// Хешируем CVV
	hashedCVV, err := bcrypt.GenerateFromPassword([]byte(card.CVV), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	card.CVV = string(hashedCVV)

	// Вычисляем HMAC для card_number + expiry_date
	data := card.CardNumber + card.ExpiryDate
	h := hmac.New(sha256.New, []byte(s.hmacSecret))
	h.Write([]byte(data))
	card.HMAC = hex.EncodeToString(h.Sum(nil))

	// Сохраняем карту в базе
	if err := s.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}

	return card, nil
}

func (s *cardService) GetCards(ctx context.Context, accountID int64, userID int64) ([]*models.Card, error) {
	// Проверяем, существует ли счет и принадлежит ли он пользователю
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

	// Получаем карты
	cards, err := s.cardRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Проверяем HMAC для каждой карты
	for _, card := range cards {
		data := card.CardNumber + card.ExpiryDate
		h := hmac.New(sha256.New, []byte(s.hmacSecret))
		h.Write([]byte(data))
		expectedHMAC := hex.EncodeToString(h.Sum(nil))
		if !hmac.Equal([]byte(card.HMAC), []byte(expectedHMAC)) {
			return nil, errors.New("HMAC verification failed for card ID " + string(card.ID))
		}
	}

	return cards, nil
}
