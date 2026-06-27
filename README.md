# airbar-finance

Go finance service for Airbar (Scenario B): ledger, escrow, Zibal PSP, wallet, withdrawal.

**Status:** F0–F3 done — bootstrap, ledger SSOT, wallet accounts, escrow lifecycle + gRPC.

Full docs: [docs/README.md](docs/README.md) · [docs/development.md](docs/development.md)

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

**Setup, env, Docker, migrations, tests:** [docs/development.md](docs/development.md)  
**Phase reports:** [docs/README.md](docs/README.md)

## CI

- `ci.yml` — `go mod verify`, `go vet`, `go test`, `go build`
- `notify-events.yml` — Telegram repo notifications
