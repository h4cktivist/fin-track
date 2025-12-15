package kafka

import (
	"context"

	"fin-track-app/internal/domain"
)

type EventPublisher interface {
	PublishTransactions(ctx context.Context, msg domain.TransactionMessage) error
}
