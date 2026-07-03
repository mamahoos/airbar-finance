-- +goose Up
-- F11 foundation: promo credit accounts and grant registry (balance from ledger SSOT).
CREATE TABLE finance.credit_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'IRT',
    account_code TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT credit_accounts_user_currency_unique UNIQUE (user_id, currency),
    CONSTRAINT credit_accounts_account_code_unique UNIQUE (account_code)
);

CREATE INDEX idx_credit_accounts_user_id ON finance.credit_accounts (user_id);

CREATE TABLE finance.credit_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    amount_rials BIGINT NOT NULL,
    reason TEXT NOT NULL,
    campaign_ref TEXT,
    expires_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    granted_by TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    reversed_at TIMESTAMPTZ,
    reverse_reason TEXT,
    reversed_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT credit_grants_amount_positive CHECK (amount_rials > 0),
    CONSTRAINT credit_grants_idempotency_key_unique UNIQUE (idempotency_key)
);

CREATE INDEX idx_credit_grants_user_id ON finance.credit_grants (user_id);
CREATE INDEX idx_credit_grants_status ON finance.credit_grants (status);

-- +goose Down
DROP TABLE IF EXISTS finance.credit_grants;
DROP TABLE IF EXISTS finance.credit_accounts;
