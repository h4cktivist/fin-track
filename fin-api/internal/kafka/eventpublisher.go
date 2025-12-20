package kafka

import (
	"context"

	"fin-api/internal/domain"
)

type EventPublisher interface {
	PublishTransactions(ctx context.Context, msg domain.TransactionMessage) error
}
