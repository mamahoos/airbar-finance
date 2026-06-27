-- +goose Up
-- F3: escrow aggregate — lifecycle metadata; balances come from ledger_entries SSOT.
CREATE TABLE finance.escrows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shipment_id TEXT NOT NULL,
    carrier_user_id TEXT NOT NULL,
    payer_user_id TEXT NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status TEXT NOT NULL,
    payment_order_id TEXT,
    funding_source TEXT,
    funded_at TIMESTAMPTZ,
    released_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT escrows_shipment_id_unique UNIQUE (shipment_id)
);

CREATE INDEX idx_escrows_status ON finance.escrows (status);
CREATE INDEX idx_escrows_payer_user_id ON finance.escrows (payer_user_id);

-- +goose Down
DROP TABLE IF EXISTS finance.escrows;
