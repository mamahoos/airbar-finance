-- +goose Up
-- Track promo credit portion of escrow funding for non-withdrawable refund routing.
ALTER TABLE finance.escrows
    ADD COLUMN promo_credit_funded BIGINT NOT NULL DEFAULT 0 CHECK (promo_credit_funded >= 0);

-- +goose Down
ALTER TABLE finance.escrows DROP COLUMN IF EXISTS promo_credit_funded;
