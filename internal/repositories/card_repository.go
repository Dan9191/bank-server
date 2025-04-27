package repositories

import (
	"context"
	"database/sql"

	"github.com/bank-service/internal/models"
)

type cardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) CardRepository {
	return &cardRepository{db: db}
}

func (r *cardRepository) Create(ctx context.Context, card *models.Card) error {
	query := `
		INSERT INTO bank.cards (account_id, card_number, expiry_date, cvv, hmac, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	err := r.db.QueryRowContext(ctx, query,
		card.AccountID,
		card.CardNumber,
		card.ExpiryDate,
		card.CVV,
		card.HMAC,
		card.CreatedAt,
		card.UpdatedAt,
	).Scan(&card.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *cardRepository) FindByAccountID(ctx context.Context, accountID int64) ([]*models.Card, error) {
	query := `
		SELECT id, account_id, card_number, expiry_date, cvv, hmac, created_at, updated_at
		FROM bank.cards
		WHERE account_id = $1`
	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		card := &models.Card{}
		if err := rows.Scan(&card.ID, &card.AccountID, &card.CardNumber, &card.ExpiryDate, &card.CVV, &card.HMAC, &card.CreatedAt, &card.UpdatedAt); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cards, nil
}
