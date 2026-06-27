-- +goose Up
-- F9: audit trail for finance aggregate state changes.
CREATE TABLE finance.finance_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_finance_events_aggregate ON finance.finance_events (aggregate_type, aggregate_id);
CREATE INDEX idx_finance_events_event_type ON finance.finance_events (event_type);
CREATE INDEX idx_finance_events_created_at ON finance.finance_events (created_at);

-- +goose Down
DROP TABLE IF EXISTS finance.finance_events;
