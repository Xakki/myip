package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"myip/internal/rdap"
)

const (
	cacheTTL     = 7 * 24 * time.Hour
	refreshAfter = 24 * time.Hour
)

type cachedRDAP struct {
	FetchedAt time.Time `json:"fetched_at"`
	Info      rdap.Info `json:"info"`
}

// RedisStore provides RDAP cache and counters stored in Redis.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore initializes Redis access.
func NewRedisStore(addr, user, password string) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: user,
		Password: password,
	})
	return &RedisStore{client: client}
}

// Ping verifies Redis connectivity.
func (s *RedisStore) Ping(ctx context.Context) error {
	if err := s.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	return nil
}

// GetCached returns cached RDAP info if present.
func (s *RedisStore) GetCached(ctx context.Context, ip string) (rdap.Info, time.Time, bool, error) {
	key := cacheKey(ip)
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return rdap.Info{}, time.Time{}, false, nil
		}
		return rdap.Info{}, time.Time{}, false, fmt.Errorf("get cache: %w", err)
	}

	var cached cachedRDAP
	if err := json.Unmarshal([]byte(data), &cached); err != nil {
		return rdap.Info{}, time.Time{}, false, fmt.Errorf("decode cache: %w", err)
	}

	return cached.Info, cached.FetchedAt, true, nil
}

// SetCached stores RDAP data with TTL.
func (s *RedisStore) SetCached(ctx context.Context, ip string, info rdap.Info, fetchedAt time.Time) error {
	cached := cachedRDAP{FetchedAt: fetchedAt, Info: info}
	payload, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("encode cache: %w", err)
	}

	if err := s.client.Set(ctx, cacheKey(ip), payload, cacheTTL).Err(); err != nil {
		return fmt.Errorf("set cache: %w", err)
	}
	return nil
}

// NeedsRefresh reports whether cached data should be refreshed.
func NeedsRefresh(fetchedAt time.Time) bool {
	if fetchedAt.IsZero() {
		return true
	}
	return time.Since(fetchedAt) >= refreshAfter
}

// IncrementCount increases the counter for the IP and returns the new value.
func (s *RedisStore) IncrementCount(ctx context.Context, ip string) (int64, error) {
	count, err := s.client.Incr(ctx, countKey(ip)).Result()
	if err != nil {
		return 0, fmt.Errorf("increment count: %w", err)
	}
	return count, nil
}

func cacheKey(ip string) string {
	return "rdap:" + ip
}

func countKey(ip string) string {
	return "count:" + ip
}
