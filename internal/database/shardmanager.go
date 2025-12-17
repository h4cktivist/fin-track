package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"fin-track-app/internal/config"
)

type IntShardManager struct {
	shards     []*pgxpool.Pool
	shardCount int
	mu         sync.RWMutex
}

func NewIntShardManager(ctx context.Context, cfg config.PostgresConfig) (*IntShardManager, error) {
	if len(cfg.Shards) == 0 {
		return nil, fmt.Errorf("no shards configured")
	}

	manager := &IntShardManager{
		shards:     make([]*pgxpool.Pool, len(cfg.Shards)),
		shardCount: len(cfg.Shards),
	}

	for i, shardCfg := range cfg.Shards {
		pool, err := newShardPool(ctx, shardCfg.ConnURL)
		if err != nil {
			for j := 0; j < i; j++ {
				manager.shards[j].Close()
			}
			return nil, fmt.Errorf("create shard %d: %w", i, err)
		}
		manager.shards[i] = pool
	}

	return manager, nil
}

func newShardPool(ctx context.Context, connURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connURL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	return pool, nil
}

func (sm *IntShardManager) GetShardForUser(userID int) *pgxpool.Pool {
	shardIndex := sm.GetShardIndex(userID)
	return sm.shards[shardIndex]
}

func (sm *IntShardManager) GetShardIndex(userID int) int {
	if userID < 0 {
		userID = -userID
	}
	return userID % sm.shardCount
}

func (sm *IntShardManager) GetAllShards() []*pgxpool.Pool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	shards := make([]*pgxpool.Pool, sm.shardCount)
	copy(shards, sm.shards)
	return shards
}

func (sm *IntShardManager) GetShardByIndex(index int) *pgxpool.Pool {
	if index < 0 || index >= sm.shardCount {
		return nil
	}
	return sm.shards[index]
}

func (sm *IntShardManager) Close() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, pool := range sm.shards {
		if pool != nil {
			pool.Close()
		}
	}
}

func (sm *IntShardManager) GetShardsForUsers(userIDs []int) map[int]*pgxpool.Pool {
	result := make(map[int]*pgxpool.Pool)
	for _, userID := range userIDs {
		result[userID] = sm.GetShardForUser(userID)
	}
	return result
}
