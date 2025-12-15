package grpcclient

import (
	"context"
	"fin-track-app/internal/domain"
)

type TransactionClient interface {
	FetchTransactions(ctx context.Context, userID string) ([]domain.Transaction, error)
}
