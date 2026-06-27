package idempotency

import (
	"context"

	domainidempotency "github.com/mamahoos/airbar-finance/internal/domain/idempotency"
)

// SnapshotCache caches completed idempotency responses.
type SnapshotCache interface {
	Get(ctx context.Context, idempotencyKey string) (map[string]any, bool, error)
	Set(ctx context.Context, idempotencyKey string, snapshot map[string]any) error
}

// Guard coordinates Redis + DB idempotency dedup for mutating RPCs.
type Guard struct {
	repo  domainidempotency.Repository
	cache SnapshotCache
}

// NewGuard creates an idempotency guard.
func NewGuard(repo domainidempotency.Repository, cache SnapshotCache) *Guard {
	return &Guard{repo: repo, cache: cache}
}

// Begin resolves an idempotency key before handler execution.
// A non-nil replay snapshot means the handler should not run.
func (g *Guard) Begin(ctx context.Context, key, scope, resourceType, resourceID string) (map[string]any, error) {
	if key == "" {
		return nil, domainidempotency.ErrKeyRequired
	}

	if g.cache != nil {
		snapshot, ok, err := g.cache.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		if ok {
			return snapshot, nil
		}
	}

	acquired, existing, err := g.repo.TryBeginProcessing(ctx, &domainidempotency.Record{
		Key:          key,
		Scope:        scope,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
	if err != nil {
		return nil, err
	}
	if !acquired {
		switch existing.Status {
		case domainidempotency.StatusCompleted:
			if g.cache != nil && existing.ResponseSnapshot != nil {
				_ = g.cache.Set(ctx, key, existing.ResponseSnapshot)
			}
			return existing.ResponseSnapshot, nil
		case domainidempotency.StatusProcessing:
			return nil, domainidempotency.ErrConflict
		default:
			return nil, domainidempotency.ErrConflict
		}
	}
	return nil, nil
}

// Complete stores a successful handler snapshot.
func (g *Guard) Complete(ctx context.Context, key string, snapshot map[string]any) error {
	if err := g.repo.Complete(ctx, key, snapshot); err != nil {
		return err
	}
	if g.cache != nil {
		return g.cache.Set(ctx, key, snapshot)
	}
	return nil
}

// Rollback removes an in-flight record after handler failure.
func (g *Guard) Rollback(ctx context.Context, key string) error {
	return g.repo.DeleteProcessing(ctx, key)
}
