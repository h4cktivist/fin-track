package grpcclient

import (
	"context"
	"fin-analytics/internal/domain"
)

type TransactionClient interface {
	FetchTransactions(ctx context.Context, userID int) ([]domain.Transaction, error)
}
