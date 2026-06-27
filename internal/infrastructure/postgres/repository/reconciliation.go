package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainrecon "github.com/mamahoos/airbar-finance/internal/domain/reconciliation"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
)

// ReconciliationRepository implements domainrecon.Repository using PostgreSQL.
type ReconciliationRepository struct {
	pool *pgxpool.Pool
}

// NewReconciliationRepository creates a Postgres-backed reconciliation repository.
func NewReconciliationRepository(pool *pgxpool.Pool) *ReconciliationRepository {
	return &ReconciliationRepository{pool: pool}
}

// Create inserts a reconciliation run row.
func (r *ReconciliationRepository) Create(ctx context.Context, run *domainrecon.Run) error {
	id := uuid.NewString()
	findingsJSON, err := json.Marshal(run.Findings)
	if err != nil {
		return err
	}

	var startedAt time.Time
	err = r.querier(ctx).QueryRow(ctx, `
		INSERT INTO finance.reconciliation_runs (id, status, findings, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING started_at
	`, id, string(run.Status), findingsJSON, run.StartedAt, run.CompletedAt).Scan(&startedAt)
	if err != nil {
		return err
	}

	run.ID = id
	run.StartedAt = startedAt
	return nil
}

// GetByID loads a reconciliation run by id.
func (r *ReconciliationRepository) GetByID(ctx context.Context, id string) (*domainrecon.Run, error) {
	row := r.querier(ctx).QueryRow(ctx, reconciliationSelectSQL+` WHERE id = $1`, id)
	run, err := scanReconciliationRun(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainrecon.ErrNotFound
	}
	return run, err
}

// List returns reconciliation runs ordered by started_at descending.
func (r *ReconciliationRepository) List(ctx context.Context) ([]domainrecon.Run, error) {
	rows, err := r.querier(ctx).Query(ctx, reconciliationSelectSQL+` ORDER BY started_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []domainrecon.Run
	for rows.Next() {
		run, err := scanReconciliationRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *run)
	}
	return runs, rows.Err()
}

const reconciliationSelectSQL = `
	SELECT id, status, findings, started_at, completed_at
	FROM finance.reconciliation_runs
`

type reconciliationScanner interface {
	Scan(dest ...any) error
}

func scanReconciliationRun(row reconciliationScanner) (*domainrecon.Run, error) {
	var run domainrecon.Run
	var status string
	var findingsJSON []byte
	var completedAt *time.Time

	err := row.Scan(&run.ID, &status, &findingsJSON, &run.StartedAt, &completedAt)
	if err != nil {
		return nil, err
	}

	run.Status = domainrecon.Status(status)
	run.CompletedAt = completedAt
	if len(findingsJSON) > 0 {
		if err := json.Unmarshal(findingsJSON, &run.Findings); err != nil {
			return nil, err
		}
	}
	return &run, nil
}

func (r *ReconciliationRepository) querier(ctx context.Context) pgxQuerier {
	if tx, ok := pg.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}
