-- +goose Up
-- F4: payment orders for Zibal DIRECT shipment pay and wallet topup.
CREATE TABLE finance.payment_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shipment_id TEXT,
    payer_user_id TEXT NOT NULL,
    purpose TEXT NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status TEXT NOT NULL,
    authority TEXT,
    redirect_url TEXT,
    success_url TEXT NOT NULL,
    failure_url TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    agreed_price BIGINT,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_orders_authority ON finance.payment_orders (authority);
CREATE INDEX idx_payment_orders_shipment_id ON finance.payment_orders (shipment_id);
CREATE INDEX idx_payment_orders_payer_user_id ON finance.payment_orders (payer_user_id);

-- +goose Down
DROP TABLE IF EXISTS finance.payment_orders;
