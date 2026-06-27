package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

// NewClient creates a Redis client from a URL (redis://host:port/db).
func NewClient(redisURL string) (*goredis.Client, error) {
	opts, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return goredis.NewClient(opts), nil
}

// Ping verifies connectivity to Redis.
func Ping(ctx context.Context, client *goredis.Client) error {
	return client.Ping(ctx).Err()
}
