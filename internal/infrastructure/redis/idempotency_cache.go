package redis

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const idempotencyKeyPrefix = "idempotency:"

// IdempotencyCache caches completed idempotency snapshots in Redis.
type IdempotencyCache struct {
	client *goredis.Client
	ttl    time.Duration
}

// NewIdempotencyCache creates a Redis-backed idempotency cache (24h TTL per F8.3).
func NewIdempotencyCache(client *goredis.Client) *IdempotencyCache {
	return &IdempotencyCache{
		client: client,
		ttl:    24 * time.Hour,
	}
}

// Get returns a cached snapshot map or nil if missing.
func (c *IdempotencyCache) Get(ctx context.Context, idempotencyKey string) (map[string]any, bool, error) {
	raw, err := c.client.Get(ctx, idempotencyKeyPrefix+idempotencyKey).Bytes()
	if err == goredis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var snapshot map[string]any
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, false, err
	}
	return snapshot, true, nil
}

// Set stores a completed snapshot with TTL.
func (c *IdempotencyCache) Set(ctx context.Context, idempotencyKey string, snapshot map[string]any) error {
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, idempotencyKeyPrefix+idempotencyKey, raw, c.ttl).Err()
}
