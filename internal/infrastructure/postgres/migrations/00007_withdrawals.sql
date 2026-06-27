-- +goose Up
-- F6: withdrawal requests — destination_hash only, no plain IBAN storage.
CREATE TABLE finance.withdrawals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status TEXT NOT NULL,
    destination_hash TEXT NOT NULL,
    provider_ref TEXT,
    reject_reason TEXT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_withdrawals_user_id ON finance.withdrawals (user_id);
CREATE INDEX idx_withdrawals_status ON finance.withdrawals (status);

-- +goose Down
DROP TABLE IF EXISTS finance.withdrawals;
