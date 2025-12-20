package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"fin-analytics/internal/domain"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(client *redis.Client, ttl time.Duration) *Cache {
	return &Cache{client: client, ttl: ttl}
}

func (c *Cache) key(userID int) string {
	return fmt.Sprintf("fintrack:stats:%d", userID)
}

func (c *Cache) Get(ctx context.Context, userID int) (*domain.FinanceStats, error) {
	data, err := c.client.Get(ctx, c.key(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var stats domain.FinanceStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("unmarshal stats: %w", err)
	}
	return &stats, nil
}

func (c *Cache) Set(ctx context.Context, stats domain.FinanceStats) error {
	payload, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("marshal stats: %w", err)
	}
	if err := c.client.Set(ctx, c.key(stats.UserID), payload, c.ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}
