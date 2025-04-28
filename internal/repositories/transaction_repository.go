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

func (r *transactionRepository) FindByAccountID(ctx context.Context, accountID int64) ([]*models.Transaction, error) {
	query := `
		SELECT id, account_id, amount, type, description, created_at
		FROM bank.transactions
		WHERE account_id = $1
		ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		transaction := &models.Transaction{}
		if err := rows.Scan(&transaction.ID, &transaction.AccountID, &transaction.Amount, &transaction.Type, &transaction.Description, &transaction.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}
