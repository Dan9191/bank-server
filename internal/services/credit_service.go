package services

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/bank-service/internal/models"
	"github.com/bank-service/internal/repositories"
)

type creditService struct {
	creditRepo repositories.CreditRepository
	userRepo   repositories.UserRepository
}

func NewCreditService(creditRepo repositories.CreditRepository, userRepo repositories.UserRepository) CreditService {
	return &creditService{
		creditRepo: creditRepo,
		userRepo:   userRepo,
	}
}

func (s *creditService) CreateCredit(ctx context.Context, userID int64, amount, interestRate float64, termMonths int) (*models.Credit, error) {
	// Проверяем, существует ли пользователь
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Создаём кредит
	credit := &models.Credit{
		UserID:       userID,
		Amount:       amount,
		InterestRate: interestRate,
		TermMonths:   termMonths,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Валидируем кредит
	if err := credit.Validate(); err != nil {
		return nil, err
	}

	// Сохраняем кредит
	if err := s.creditRepo.CreateCredit(ctx, credit); err != nil {
		return nil, err
	}

	// Вычисляем аннуитетный платёж
	monthlyRate := interestRate / 100 / 12
	annuityFactor := (monthlyRate * math.Pow(1+monthlyRate, float64(termMonths))) / (math.Pow(1+monthlyRate, float64(termMonths)) - 1)
	monthlyPayment := amount * annuityFactor

	// Создаём график платежей
	currentDate := time.Now().AddDate(0, 1, 0) // Первый платёж через месяц
	for i := 0; i < termMonths; i++ {
		paymentSchedule := &models.PaymentSchedule{
			CreditID:    credit.ID,
			PaymentDate: currentDate.AddDate(0, i, 0),
			Amount:      math.Round(monthlyPayment*100) / 100,
			Paid:        false,
			Penalty:     0.0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := paymentSchedule.Validate(); err != nil {
			return nil, err
		}

		if err := s.creditRepo.CreatePaymentSchedule(ctx, paymentSchedule); err != nil {
			return nil, err
		}
	}

	return credit, nil
}

func (s *creditService) GetCredits(ctx context.Context, userID int64) ([]*models.Credit, error) {
	// Проверяем, существует ли пользователь
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Получаем кредиты
	credits, err := s.creditRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return credits, nil
}

func (s *creditService) GetPaymentSchedules(ctx context.Context, creditID, userID int64) ([]*models.PaymentSchedule, error) {
	// Проверяем, существует ли кредит и принадлежит ли он пользователю
	credits, err := s.creditRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var creditExists bool
	for _, credit := range credits {
		if credit.ID == creditID {
			creditExists = true
			break
		}
	}
	if !creditExists {
		return nil, errors.New("credit not found or unauthorized")
	}

	// Получаем график платежей
	schedules, err := s.creditRepo.FindPaymentSchedulesByCreditID(ctx, creditID)
	if err != nil {
		return nil, err
	}
	return schedules, nil
}
