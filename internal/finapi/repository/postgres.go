package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fin-track-app/internal/domain"
)

type PostgresTransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) CreateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	query := `
		INSERT INTO transactions (user_id, amount, category, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at;
	`

	row := r.db.QueryRow(ctx, query, tx.UserID, tx.Amount, tx.Category, tx.Type)
	if err := row.Scan(&tx.ID, &tx.CreatedAt); err != nil {
		return domain.Transaction{}, fmt.Errorf("insert transaction: %w", err)
	}

	return tx, nil
}

func (r *PostgresTransactionRepository) ListUserTransactions(ctx context.Context, userID string) ([]domain.Transaction, error) {
	query := `
		SELECT id, user_id, amount, category, type, created_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at ASC;
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var result []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.Amount, &tx.Category, &tx.Type, &tx.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		result = append(result, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return result, nil
}

func (r *PostgresTransactionRepository) UpdateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	query := `
		UPDATE transactions
		SET amount = $1,
		    category = $2,
		    type = $3
		WHERE id = $4 AND user_id = $5
		RETURNING created_at;
	`

	row := r.db.QueryRow(ctx, query, tx.Amount, tx.Category, tx.Type, tx.ID, tx.UserID)
	if err := row.Scan(&tx.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Transaction{}, domain.ErrTransactionNotFound
		}
		return domain.Transaction{}, fmt.Errorf("update transaction: %w", err)
	}

	return tx, nil
}

func (r *PostgresTransactionRepository) DeleteTransaction(ctx context.Context, userID string, transactionID int64) error {
	commandTag, err := r.db.Exec(ctx, `
		DELETE FROM transactions
		WHERE id = $1 AND user_id = $2;
	`, transactionID, userID)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrTransactionNotFound
	}
	return nil
}
