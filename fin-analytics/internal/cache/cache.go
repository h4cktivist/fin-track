package cache

import (
	"context"

	"fin-analytics/internal/domain"
)

type StatsCache interface {
	Get(ctx context.Context, userID int) (*domain.FinanceStats, error)
	Set(ctx context.Context, stats domain.FinanceStats) error
}
