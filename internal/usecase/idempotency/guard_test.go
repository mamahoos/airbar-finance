package idempotency_test

import (
	"context"
	"testing"

	domainidempotency "github.com/mamahoos/airbar-finance/internal/domain/idempotency"
	idempotencyuc "github.com/mamahoos/airbar-finance/internal/usecase/idempotency"
)

type memoryRepo struct {
	records map[string]*domainidempotency.Record
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{records: make(map[string]*domainidempotency.Record)}
}

func (m *memoryRepo) TryBeginProcessing(_ context.Context, record *domainidempotency.Record) (bool, *domainidempotency.Record, error) {
	if existing, ok := m.records[record.Key]; ok {
		return false, cloneRecord(existing), nil
	}
	record.Status = domainidempotency.StatusProcessing
	m.records[record.Key] = cloneRecord(record)
	return true, record, nil
}

func (m *memoryRepo) GetByKey(_ context.Context, key string) (*domainidempotency.Record, error) {
	record, ok := m.records[key]
	if !ok {
		return nil, domainidempotency.ErrNotFound
	}
	return cloneRecord(record), nil
}

func (m *memoryRepo) Complete(_ context.Context, key string, snapshot map[string]any) error {
	record, ok := m.records[key]
	if !ok || record.Status != domainidempotency.StatusProcessing {
		return domainidempotency.ErrNotFound
	}
	record.Status = domainidempotency.StatusCompleted
	record.ResponseSnapshot = snapshot
	return nil
}

func (m *memoryRepo) DeleteProcessing(_ context.Context, key string) error {
	record, ok := m.records[key]
	if !ok || record.Status != domainidempotency.StatusProcessing {
		return nil
	}
	delete(m.records, key)
	return nil
}

type memoryCache struct {
	values map[string]map[string]any
}

func newMemoryCache() *memoryCache {
	return &memoryCache{values: make(map[string]map[string]any)}
}

func (c *memoryCache) Get(_ context.Context, key string) (map[string]any, bool, error) {
	value, ok := c.values[key]
	return value, ok, nil
}

func (c *memoryCache) Set(_ context.Context, key string, snapshot map[string]any) error {
	c.values[key] = snapshot
	return nil
}

func cloneRecord(record *domainidempotency.Record) *domainidempotency.Record {
	copyRecord := *record
	if record.ResponseSnapshot != nil {
		copyRecord.ResponseSnapshot = make(map[string]any, len(record.ResponseSnapshot))
		for k, v := range record.ResponseSnapshot {
			copyRecord.ResponseSnapshot[k] = v
		}
	}
	return &copyRecord
}

func TestGuardBeginRequiresKey(t *testing.T) {
	guard := idempotencyuc.NewGuard(newMemoryRepo(), newMemoryCache())

	_, err := guard.Begin(context.Background(), "", "escrow.create", "shipment", "sh-1")
	if err == nil || !domainidempotency.IsValidation(err) {
		t.Fatalf("Begin() error = %v, want validation", err)
	}
}

func TestGuardBeginCompleteReplay(t *testing.T) {
	repo := newMemoryRepo()
	cache := newMemoryCache()
	guard := idempotencyuc.NewGuard(repo, cache)
	ctx := context.Background()

	replay, err := guard.Begin(ctx, "key-1", "escrow.create", "shipment", "sh-1")
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	if replay != nil {
		t.Fatal("expected no replay on first Begin")
	}

	snapshot := map[string]any{"status": "CREATED"}
	if err := guard.Complete(ctx, "key-1", snapshot); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	replay, err = guard.Begin(ctx, "key-1", "escrow.create", "shipment", "sh-1")
	if err != nil {
		t.Fatalf("Begin() replay error = %v", err)
	}
	if replay["status"] != "CREATED" {
		t.Fatalf("replay = %#v, want CREATED", replay)
	}
}

func TestGuardBeginConflictWhileProcessing(t *testing.T) {
	repo := newMemoryRepo()
	guard := idempotencyuc.NewGuard(repo, nil)
	ctx := context.Background()

	if _, err := guard.Begin(ctx, "key-2", "escrow.create", "shipment", "sh-2"); err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	_, err := guard.Begin(ctx, "key-2", "escrow.create", "shipment", "sh-2")
	if err == nil || !domainidempotency.IsConflict(err) {
		t.Fatalf("Begin() error = %v, want conflict", err)
	}
}

func TestGuardRollbackAllowsRetry(t *testing.T) {
	repo := newMemoryRepo()
	guard := idempotencyuc.NewGuard(repo, nil)
	ctx := context.Background()

	if _, err := guard.Begin(ctx, "key-3", "escrow.create", "shipment", "sh-3"); err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	if err := guard.Rollback(ctx, "key-3"); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	replay, err := guard.Begin(ctx, "key-3", "escrow.create", "shipment", "sh-3")
	if err != nil {
		t.Fatalf("Begin() after rollback error = %v", err)
	}
	if replay != nil {
		t.Fatal("expected fresh Begin after rollback")
	}
}
