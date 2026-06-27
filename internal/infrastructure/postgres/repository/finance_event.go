package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	domainaudit "github.com/mamahoos/airbar-finance/internal/domain/audit"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
)

// FinanceEventRepository implements domainaudit.Repository using PostgreSQL.
type FinanceEventRepository struct {
	pool *pgxpool.Pool
}

// NewFinanceEventRepository creates a Postgres-backed finance event repository.
func NewFinanceEventRepository(pool *pgxpool.Pool) *FinanceEventRepository {
	return &FinanceEventRepository{pool: pool}
}

// Create inserts a finance audit event.
func (r *FinanceEventRepository) Create(ctx context.Context, event *domainaudit.Event) error {
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	id := uuid.NewString()
	var createdAt time.Time
	err = r.querier(ctx).QueryRow(ctx, `
		INSERT INTO finance.finance_events (id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`, id, event.AggregateType, event.AggregateID, string(event.EventType), payloadJSON).
		Scan(&createdAt)
	if err != nil {
		return err
	}

	event.ID = id
	event.CreatedAt = createdAt
	return nil
}

// CountByAggregate returns how many events exist for an aggregate.
func (r *FinanceEventRepository) CountByAggregate(ctx context.Context, aggregateType, aggregateID string) (int64, error) {
	var count int64
	err := r.querier(ctx).QueryRow(ctx, `
		SELECT COUNT(*) FROM finance.finance_events
		WHERE aggregate_type = $1 AND aggregate_id = $2
	`, aggregateType, aggregateID).Scan(&count)
	return count, err
}

func (r *FinanceEventRepository) querier(ctx context.Context) pgxQuerier {
	if tx, ok := pg.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}
