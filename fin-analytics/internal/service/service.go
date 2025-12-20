package service

import (
	"context"
	"encoding/json"
	"fin-analytics/internal/cache"
	client "fin-analytics/internal/grpcclient"
	"fin-analytics/internal/statscalculator"
	"fmt"

	"github.com/IBM/sarama"

	"fin-analytics/internal/domain"
)

type Service struct {
	cache  cache.StatsCache
	client client.TransactionClient
}

func New(cache cache.StatsCache, client client.TransactionClient) *Service {
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

	stats := statscalculator.CalculateStats(payload.Transactions)
	if err := s.cache.Set(ctx, stats); err != nil {
		return fmt.Errorf("cache stats: %w", err)
	}
	return nil
}

func (s *Service) GetStats(ctx context.Context, userID int) (domain.FinanceStats, error) {
	if cached, err := s.cache.Get(ctx, userID); err == nil && cached != nil {
		return *cached, nil
	}

	txs, err := s.client.FetchTransactions(ctx, userID)
	if err != nil {
		return domain.FinanceStats{}, err
	}
	stats := statscalculator.CalculateStats(txs)
	if err := s.cache.Set(ctx, stats); err != nil {
		return domain.FinanceStats{}, err
	}
	return stats, nil
}
