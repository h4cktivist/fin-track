package service

import (
	"context"

	"fin-track-app/internal/domain"
)

type StatsCache interface {
	Get(ctx context.Context, userID string) (*domain.FinanceStats, error)
	Set(ctx context.Context, stats domain.FinanceStats) error
}

type TransactionClient interface {
	FetchTransactions(ctx context.Context, userID string) ([]domain.Transaction, error)
}
