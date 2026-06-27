-- +goose Up
-- F8: idempotency dedup for mutating gRPC commands.
CREATE TABLE finance.idempotency_records (
    idempotency_key TEXT PRIMARY KEY,
    scope TEXT NOT NULL,
    resource_type TEXT,
    resource_id TEXT,
    status TEXT NOT NULL,
    response_snapshot JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_idempotency_records_scope ON finance.idempotency_records (scope);
CREATE INDEX idx_idempotency_records_status ON finance.idempotency_records (status);

-- +goose Down
DROP TABLE IF EXISTS finance.idempotency_records;
