package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainidempotency "github.com/mamahoos/airbar-finance/internal/domain/idempotency"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
)

// IdempotencyRepository implements domainidempotency.Repository using PostgreSQL.
type IdempotencyRepository struct {
	pool *pgxpool.Pool
}

// NewIdempotencyRepository creates a Postgres-backed idempotency repository.
func NewIdempotencyRepository(pool *pgxpool.Pool) *IdempotencyRepository {
	return &IdempotencyRepository{pool: pool}
}

// TryBeginProcessing inserts PROCESSING or returns the existing row.
func (r *IdempotencyRepository) TryBeginProcessing(ctx context.Context, record *domainidempotency.Record) (bool, *domainidempotency.Record, error) {
	q := r.querier(ctx)

	var createdAt time.Time
	err := q.QueryRow(ctx, `
		INSERT INTO finance.idempotency_records (idempotency_key, scope, resource_type, resource_id, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING created_at
	`, record.Key, record.Scope, nullString(record.ResourceType), nullString(record.ResourceID), string(domainidempotency.StatusProcessing)).
		Scan(&createdAt)
	if err == nil {
		record.Status = domainidempotency.StatusProcessing
		record.CreatedAt = createdAt
		return true, record, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, nil, err
	}

	existing, err := r.GetByKey(ctx, record.Key)
	if err != nil {
		return false, nil, err
	}
	return false, existing, nil
}

// GetByKey loads an idempotency record by key.
func (r *IdempotencyRepository) GetByKey(ctx context.Context, key string) (*domainidempotency.Record, error) {
	row := r.querier(ctx).QueryRow(ctx, `
		SELECT idempotency_key, scope, resource_type, resource_id, status, response_snapshot, created_at, completed_at
		FROM finance.idempotency_records
		WHERE idempotency_key = $1
	`, key)

	var record domainidempotency.Record
	var status string
	var resourceType *string
	var resourceID *string
	var snapshotJSON []byte
	var completedAt *time.Time

	err := row.Scan(
		&record.Key,
		&record.Scope,
		&resourceType,
		&resourceID,
		&status,
		&snapshotJSON,
		&record.CreatedAt,
		&completedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainidempotency.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	record.Status = domainidempotency.Status(status)
	if resourceType != nil {
		record.ResourceType = *resourceType
	}
	if resourceID != nil {
		record.ResourceID = *resourceID
	}
	record.CompletedAt = completedAt
	if len(snapshotJSON) > 0 {
		if err := json.Unmarshal(snapshotJSON, &record.ResponseSnapshot); err != nil {
			return nil, err
		}
	}
	return &record, nil
}

// Complete marks a record completed with a response snapshot.
func (r *IdempotencyRepository) Complete(ctx context.Context, key string, snapshot map[string]any) error {
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	tag, err := r.querier(ctx).Exec(ctx, `
		UPDATE finance.idempotency_records
		SET status = $2,
		    response_snapshot = $3,
		    completed_at = now()
		WHERE idempotency_key = $1 AND status = $4
	`, key, string(domainidempotency.StatusCompleted), snapshotJSON, string(domainidempotency.StatusProcessing))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domainidempotency.ErrNotFound
	}
	return nil
}

// DeleteProcessing removes an in-flight record after handler failure.
func (r *IdempotencyRepository) DeleteProcessing(ctx context.Context, key string) error {
	_, err := r.querier(ctx).Exec(ctx, `
		DELETE FROM finance.idempotency_records
		WHERE idempotency_key = $1 AND status = $2
	`, key, string(domainidempotency.StatusProcessing))
	return err
}

func (r *IdempotencyRepository) querier(ctx context.Context) pgxQuerier {
	if tx, ok := pg.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func nullString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
