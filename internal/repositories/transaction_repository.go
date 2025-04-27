package repositories

import (
	"context"
	"database/sql"

	"github.com/bank-service/internal/models"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error {
	query := `
		INSERT INTO bank.transactions (account_id, amount, type, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	err := tx.QueryRowContext(ctx, query,
		transaction.AccountID,
		transaction.Amount,
		transaction.Type,
		transaction.Description,
		transaction.CreatedAt,
	).Scan(&transaction.ID)
	if err != nil {
		return err
	}
	return nil
}
