-- +goose Up
ALTER TABLE finance.withdrawals
    ADD COLUMN IF NOT EXISTS payout_channel TEXT,
    ADD COLUMN IF NOT EXISTS receipt_url TEXT;

-- +goose Down
ALTER TABLE finance.withdrawals
    DROP COLUMN IF EXISTS receipt_url,
    DROP COLUMN IF EXISTS payout_channel;
