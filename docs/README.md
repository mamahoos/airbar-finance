# airbar-finance — Documentation

Engineering docs for the Go finance service (Scenario B).

## Roadmap & phases

| Phase | Status | Report |
|-------|--------|--------|
| F0 Bootstrap | Done | [tasks/F0-bootstrap-report.md](./tasks/F0-bootstrap-report.md) |
| F1 Ledger Core | Done | [tasks/F1-ledger-report.md](./tasks/F1-ledger-report.md) |
| F2 Wallet accounts | Planned | — |
| F3 Escrow | Planned | — |

Parent monorepo roadmap: [`../../docs/09-roadmap-هسته-مالی.md`](../../docs/09-roadmap-هسته-مالی.md) (if checked out alongside `airbar.app-staging`).

## Quick reference

| Topic | Location |
|-------|----------|
| Local setup, Docker, env | [development.md](./development.md) |
| gRPC contract | [`../proto/airbar_finance_v1.proto`](../proto/airbar_finance_v1.proto) |
| Migrations | `internal/infrastructure/postgres/migrations/` |
| Architecture | [`../README.md`](../README.md) |

## Configuration (production rule)

- **All runtime settings from environment** — no hardcoded defaults in `config.Load()`.
- **`.env.example`** — committed template; copy to `.env` for local dev only.
- **Production** — inject the same keys via orchestrator secrets (Kubernetes, ECS, etc.); never commit `.env`.

Required env keys: `DATABASE_URL`, `REDIS_URL`, `GRPC_PORT`, `HTTP_PORT`, `ZIBAL_SANDBOX`, `ZIBAL_MERCHANT`, `FINANCE_PUBLIC_BASE_URL`, `PLATFORM_FEE_PERCENT`.

## Testing

```bash
make verify                                    # go vet + test + build
make migrate-up                                # requires goose + running postgres
TEST_DATABASE_URL=... go test ./internal/infrastructure/postgres/repository/ -run Integration -v
```

See [development.md](./development.md) for full commands.
