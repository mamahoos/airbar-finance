-- +goose Up
-- F0 baseline: dedicated schema for airbar_finance. Domain tables arrive in F1+ migrations.
CREATE SCHEMA IF NOT EXISTS finance;

-- +goose Down
DROP SCHEMA IF EXISTS finance CASCADE;
