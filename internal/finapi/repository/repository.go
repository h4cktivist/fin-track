package repository

import (
	"context"

	"fin-track-app/internal/domain"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)
	ListUserTransactions(ctx context.Context, userID int) ([]domain.Transaction, error)
	UpdateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)
	DeleteTransaction(ctx context.Context, userID int, transactionID int64) error
}
