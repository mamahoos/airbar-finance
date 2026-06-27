-- +goose Up
-- F4: audit log for PSP interactions (Zibal request/verify/callback).
CREATE TABLE finance.provider_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payment_order_id UUID REFERENCES finance.payment_orders (id) ON DELETE SET NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    payload_hash TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT provider_events_idempotency_unique UNIQUE (idempotency_key)
);

CREATE INDEX idx_provider_events_payment_order_id ON finance.provider_events (payment_order_id);

-- +goose Down
DROP TABLE IF EXISTS finance.provider_events;
