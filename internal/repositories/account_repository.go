package repositories

import (
	"context"
	"database/sql"

	"github.com/bank-service/internal/models"
)

type accountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(ctx context.Context, account *models.Account) error {
	query := `
		INSERT INTO bank.accounts (user_id, balance, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	err := r.db.QueryRowContext(ctx, query,
		account.UserID,
		account.Balance,
		account.Currency,
		account.CreatedAt,
		account.UpdatedAt,
	).Scan(&account.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *accountRepository) FindByID(ctx context.Context, id int64) (*models.Account, error) {
	account := &models.Account{}
	query := `
		SELECT id, user_id, balance, currency, created_at, updated_at
		FROM bank.accounts
		WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.Balance,
		&account.Currency,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *accountRepository) FindByUserID(ctx context.Context, userID int64) ([]*models.Account, error) {
	query := `
		SELECT id, user_id, balance, currency, created_at, updated_at
		FROM bank.accounts
		WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*models.Account
	for rows.Next() {
		account := &models.Account{}
		if err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Balance,
			&account.Currency,
			&account.CreatedAt,
			&account.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (r *accountRepository) UpdateBalance(ctx context.Context, accountID int64, balance float64) error {
	query := `
		UPDATE bank.accounts
		SET balance = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, balance, accountID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}