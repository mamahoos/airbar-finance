package health

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

// Checker verifies downstream dependencies required for readiness.
type Checker struct {
	db    *pgxpool.Pool
	redis *goredis.Client
}

// NewChecker creates a readiness checker for Postgres and Redis.
func NewChecker(db *pgxpool.Pool, redis *goredis.Client) *Checker {
	return &Checker{db: db, redis: redis}
}

// Ready returns true when Postgres and Redis respond to ping.
func (c *Checker) Ready(ctx context.Context) bool {
	if err := c.db.Ping(ctx); err != nil {
		return false
	}
	if err := c.redis.Ping(ctx).Err(); err != nil {
		return false
	}
	return true
}
