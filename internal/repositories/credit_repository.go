package repositories

import (
	"context"
	"database/sql"

	"github.com/bank-service/internal/models"
)

type creditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) CreditRepository {
	return &creditRepository{db: db}
}

func (r *creditRepository) CreateCredit(ctx context.Context, credit *models.Credit) error {
	query := `
		INSERT INTO bank.credits (user_id, amount, interest_rate, term_months, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	err := r.db.QueryRowContext(ctx, query,
		credit.UserID,
		credit.Amount,
		credit.InterestRate,
		credit.TermMonths,
		credit.CreatedAt,
		credit.UpdatedAt,
	).Scan(&credit.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *creditRepository) FindByUserID(ctx context.Context, userID int64) ([]*models.Credit, error) {
	query := `
		SELECT id, user_id, amount, interest_rate, term_months, created_at, updated_at
		FROM bank.credits
		WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []*models.Credit
	for rows.Next() {
		credit := &models.Credit{}
		if err := rows.Scan(&credit.ID, &credit.UserID, &credit.Amount, &credit.InterestRate, &credit.TermMonths, &credit.CreatedAt, &credit.UpdatedAt); err != nil {
			return nil, err
		}
		credits = append(credits, credit)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return credits, nil
}

func (r *creditRepository) CreatePaymentSchedule(ctx context.Context, paymentSchedule *models.PaymentSchedule) error {
	query := `
		INSERT INTO bank.payment_schedules (credit_id, payment_date, amount, paid, penalty, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	err := r.db.QueryRowContext(ctx, query,
		paymentSchedule.CreditID,
		paymentSchedule.PaymentDate,
		paymentSchedule.Amount,
		paymentSchedule.Paid,
		paymentSchedule.Penalty,
		paymentSchedule.CreatedAt,
		paymentSchedule.UpdatedAt,
	).Scan(&paymentSchedule.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *creditRepository) FindPaymentSchedulesByCreditID(ctx context.Context, creditID int64) ([]*models.PaymentSchedule, error) {
	query := `
		SELECT id, credit_id, payment_date, amount, paid, penalty, created_at, updated_at
		FROM bank.payment_schedules
		WHERE credit_id = $1`
	rows, err := r.db.QueryContext(ctx, query, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.PaymentSchedule
	for rows.Next() {
		schedule := &models.PaymentSchedule{}
		if err := rows.Scan(&schedule.ID, &schedule.CreditID, &schedule.PaymentDate, &schedule.Amount, &schedule.Paid, &schedule.Penalty, &schedule.CreatedAt, &schedule.UpdatedAt); err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return schedules, nil
}
