-- +goose Up
-- F2: wallet account registry — metadata only; balance comes from ledger_entries SSOT.
CREATE TABLE finance.wallet_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'IRT',
    account_code TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT wallet_accounts_user_currency_unique UNIQUE (user_id, currency),
    CONSTRAINT wallet_accounts_account_code_unique UNIQUE (account_code)
);

CREATE INDEX idx_wallet_accounts_user_id ON finance.wallet_accounts (user_id);

-- +goose Down
DROP TABLE IF EXISTS finance.wallet_accounts;
