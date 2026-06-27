# Development guide

## Prerequisites

- Go 1.22+
- Docker + Docker Compose
- [goose](https://github.com/pressly/goose) for migrations (optional if using `make migrate-up`)

## First-time setup

```bash
cp .env.example .env
make up                    # postgres-finance (5434) + redis (6381)
make migrate-up            # apply goose migrations
make verify                # vet, unit tests, build
```

## Docker workflows

### Dependencies only (DB + Redis)

```bash
make up
make migrate-up
go run ./cmd/server        # uses .env (localhost URLs)
```

### Full stack (app in container)

```bash
cp .env.example .env
docker build -t airbar-finance:local .
docker compose up -d
```

`docker-compose.yml` overrides `DATABASE_URL` and `REDIS_URL` for in-compose hostnames (`postgres-finance`, `redis`). Other keys come from `.env`.

### Health checks

```bash
curl -sf http://localhost:8080/health/ready
GRPC_ADDR=localhost:50051 go run ./scripts/check_ready
```

## Environment files

| File | Commit? | Purpose |
|------|---------|---------|
| `.env.example` | Yes | Template for all required variables |
| `.env` | **No** | Local overrides (gitignored) |

Production deployments must inject env via the orchestrator — not via `.env` on disk.

## Migrations

| File | Phase | Content |
|------|-------|---------|
| `00001_baseline.sql` | F0 | `finance` schema |
| `00002_ledger.sql` | F1 | `ledger_journals`, `ledger_entries` |
| `00003_wallet_accounts.sql` | F2 | `wallet_accounts` (no balance column) |
| `00004_escrows.sql` | F3 | `escrows` lifecycle metadata |

```bash
make migrate-up
make migrate-status
make migrate-down   # rollback one step
```

## Tests

| Scope | Command |
|-------|---------|
| Unit (CI) | `go test ./...` |
| Integration (ledger + Postgres) | `make test-integration` (requires `-tags=integration`) |
| Full verify | `make verify` |

## CI

GitHub Actions (`.github/workflows/ci.yml`):

- `go mod verify`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/server`

## Proto codegen

Generated stubs are committed under `internal/gen/financev1/`. Regenerate when proto changes:

```bash
make proto   # requires protoc + protoc-gen-go + protoc-gen-go-grpc
```
