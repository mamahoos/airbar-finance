# airbar-finance

Go finance service for Airbar (Scenario B): ledger, escrow, Zibal PSP, wallet, withdrawal.

**Status:** F0 bootstrap — gRPC `CheckReady`, HTTP `/health/ready`, Docker/CI green.

## Layout (Clean Architecture)

```text
cmd/server/                         # composition root / bootstrap (F0)

internal/
  domain/                           # entities, value objects, repository interfaces
    escrow/
    payment/
    ledger/
    wallet/
    withdrawal/

  usecase/                          # application business rules (interactors)
    escrow/                         # UC-01..08
    payment/                        # UC-09..13
    wallet/                         # UC-14..15
    withdrawal/                     # UC-16..19
    treasury/                       # UC-20
    reconciliation/                 # UC-21..23
    idempotency/                    # cross-cutting

  delivery/                         # primary adapters (inbound)
    grpc/handlers/
    http/                           # /health/ready, Zibal callback

  infrastructure/                   # secondary adapters (outbound)
    config/
    postgres/
      migrations/
      repository/
    redis/
    zibal/

proto/                              # gRPC contract (airbar_finance_v1.proto)
```

### Dependency rule

```text
delivery → usecase → domain ← infrastructure
```

- **domain** — no imports from other internal layers
- **usecase** — depends on domain interfaces only
- **delivery** — calls use cases; maps proto/HTTP ↔ DTOs
- **infrastructure** — implements domain/usecase ports (postgres, redis, zibal)

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
| redis             | 6381  | maps to container 6379; idempotency cache (F8+) |
| gRPC              | 50051 | `FinanceHealthService.CheckReady` |
| HTTP              | 8080  | `/health/ready` (+ Zibal callback in F4) |

## Migrations

Migrations live in `internal/infrastructure/postgres/migrations/`.

| Migration        | Phase | Purpose              |
|------------------|-------|----------------------|
| `00001_baseline` | F0    | `finance` schema     |
| ledger tables    | F1    | journals + entries   |
| wallet_accounts  | F2    | lazy wallet create   |
| escrows          | F3    | escrow state machine |
| payment_orders   | F4    | Zibal integration    |

System account codes are constants in code (F1.3), not DB seed rows.

## Next steps (F0)

1. `infrastructure/config` — env loader
2. Postgres + Redis clients in infrastructure
3. gRPC `CheckReady` + HTTP `/health/ready` in delivery
4. `cmd/server/main.go` — wire dependencies
5. Enable CI: `go vet`, `go test`, `go build`

## CI

- `ci.yml` — `go mod verify` (build/test when code lands)
- `notify-events.yml` — Telegram repo notifications
