package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisIndexJobStateStoreConfig struct {
	Addr      string
	Password  string
	DB        int
	KeyPrefix string
	TTL       time.Duration
}

type RedisIndexJobStateStore struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewRedisIndexJobStateStore(cfg RedisIndexJobStateStoreConfig) (*RedisIndexJobStateStore, error) {
	if strings.TrimSpace(cfg.Addr) == "" {
		return nil, fmt.Errorf("redis addr is required")
	}
	if strings.TrimSpace(cfg.KeyPrefix) == "" {
		return nil, fmt.Errorf("redis key prefix is required")
	}
	if cfg.DB < 0 {
		return nil, fmt.Errorf("redis db must be >= 0")
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 24 * time.Hour
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &RedisIndexJobStateStore{client: client, keyPrefix: cfg.KeyPrefix, ttl: cfg.TTL}, nil
}

func (s *RedisIndexJobStateStore) Set(ctx context.Context, result IndexJobResult) error {
	jobID := strings.TrimSpace(result.JobID)
	if jobID == "" {
		return nil
	}

	payload, err := json.Marshal(cloneJobResult(result))
	if err != nil {
		return err
	}

	return s.client.Set(ctx, s.key(jobID), payload, s.ttl).Err()
}

func (s *RedisIndexJobStateStore) Get(ctx context.Context, jobID string) (IndexJobResult, bool, error) {
	trimmedJobID := strings.TrimSpace(jobID)
	if trimmedJobID == "" {
		return IndexJobResult{}, false, nil
	}

	value, err := s.client.Get(ctx, s.key(trimmedJobID)).Bytes()
	if err == redis.Nil {
		return IndexJobResult{}, false, nil
	}
	if err != nil {
		return IndexJobResult{}, false, err
	}

	var result IndexJobResult
	if err := json.Unmarshal(value, &result); err != nil {
		return IndexJobResult{}, false, err
	}

	return cloneJobResult(result), true, nil
}

func (s *RedisIndexJobStateStore) Close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Close()
}

func (s *RedisIndexJobStateStore) key(jobID string) string {
	return s.keyPrefix + jobID
}

var _ IndexJobStateStore = (*RedisIndexJobStateStore)(nil)
