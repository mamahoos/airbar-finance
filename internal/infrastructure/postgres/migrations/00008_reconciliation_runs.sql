-- +goose Up
-- F7: reconciliation run history for ledger integrity checks.
CREATE TABLE finance.reconciliation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status TEXT NOT NULL,
    findings JSONB NOT NULL DEFAULT '{}',
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_reconciliation_runs_started_at ON finance.reconciliation_runs (started_at DESC);

-- +goose Down
DROP TABLE IF EXISTS finance.reconciliation_runs;
