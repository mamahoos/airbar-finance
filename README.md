# airbar-finance

Go finance service for Airbar (Scenario B): ledger, escrow, Zibal PSP, wallet, withdrawal.

**Status:** F0 skeleton — hexagonal layout, local stack, migrations baseline. Application code comes next.

## Layout

```text
cmd/server/              # bootstrap (F0)
internal/
  config/                # env config (F0)
  domain/                # entities + rules
  application/             # use cases
  adapters/
    postgres/migrations/ # goose migrations (F0 baseline → F1+ domain)
    postgres/repositories/
    redis/
    zibal/
    grpc/handlers/
  http/                  # /health/ready, Zibal callback
proto/                   # gRPC contract (airbar_finance_v1.proto)
```

See project docs: `docs/02-هسته-مالی-go.md` (in staging workspace).

## Prerequisites

- Go 1.22+
- Docker + Docker Compose
- [goose](https://github.com/pressly/goose) for migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

## Local stack

```bash
cp .env.example .env
make up
make migrate-up
```

| Service           | Port  | Notes                    |
|-------------------|-------|--------------------------|
| postgres-finance  | 5434  | DB `airbar_finance`      |
| redis             | 6379  | idempotency cache (F8+)  |
| gRPC (planned)    | 50051 | after F0 bootstrap       |
| HTTP (planned)    | 8080  | health + Zibal callback  |

## Migrations

Migrations live in `internal/adapters/postgres/migrations/`.

| Migration        | Phase | Purpose              |
|------------------|-------|----------------------|
| `00001_baseline` | F0    | `finance` schema     |
| ledger tables    | F1    | journals + entries   |
| wallet_accounts  | F2    | lazy wallet create   |
| escrows          | F3    | escrow state machine |
| payment_orders   | F4    | Zibal integration    |

System account codes are constants in code (F1.3), not DB seed rows.

## Next steps (F0)

1. `config` loader from env
2. Postgres + Redis clients
3. gRPC `CheckReady` + HTTP `/health/ready`
4. `cmd/server/main.go` bootstrap
5. Enable CI: `go vet`, `go test`, `go build`

## CI

- `ci.yml` — `go mod verify` (build/test when code lands)
- `notify-events.yml` — Telegram repo notifications
