package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"

	"fin-track-app/internal/domain"
	"fin-track-app/internal/finanalytics/cache"
	"fin-track-app/internal/finanalytics/grpcclient"
)

type Service struct {
	cache  *cache.Cache
	client *grpcclient.Client
}

func New(cache *cache.Cache, client *grpcclient.Client) *Service {
	return &Service{
		cache:  cache,
		client: client,
	}
}

func (s *Service) ProcessKafkaMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var payload domain.TransactionMessage
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	stats := domain.CalculateStats(payload.Transactions)
	if err := s.cache.Set(ctx, stats); err != nil {
		return fmt.Errorf("cache stats: %w", err)
	}
	return nil
}

func (s *Service) GetStats(ctx context.Context, userID string) (domain.FinanceStats, error) {
	if cached, err := s.cache.Get(ctx, userID); err == nil && cached != nil {
		return *cached, nil
	}

	txs, err := s.client.FetchTransactions(ctx, userID)
	if err != nil {
		return domain.FinanceStats{}, err
	}
	stats := domain.CalculateStats(txs)
	if err := s.cache.Set(ctx, stats); err != nil {
		return domain.FinanceStats{}, err
	}
	return stats, nil
}
