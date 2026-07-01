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

Compose uses a **base + overlay** pattern:

| Overlay | Purpose | Command |
| ------- | ------- | ------- |
| `docker-compose.dev.yml` | Local dev (build + host ports) | `make up-dev` |
| `docker-compose.staging.yml` | Staging deploy (GHCR image) | `make up-staging IMAGE_TAG=ghcr.io/...` |
| `docker-compose.prod.yml` | Production deploy (GHCR image) | `make up-prod IMAGE_TAG=ghcr.io/...` |

Host ports (dev): Postgres **5434**, Redis **6381**, HTTP **8080**, gRPC **50051**.

Staging/production share Docker networks with airbar-core:

```bash
docker network create airbar-staging   # once on server
docker network create airbar-prod      # once on server
```

### Dependencies only (DB + Redis)

```bash
make up
make migrate-up
go run ./cmd/server        # uses .env (localhost URLs)
```

### Full stack (app in container)

```bash
cp .env.example .env
make up-dev
```

Overlays merge with `docker-compose.yml` and override `DATABASE_URL` / `REDIS_URL` for in-compose hostnames (`postgres-finance`, `redis`). App-specific keys come from `.env`, `.env.staging`, or `.env.production`.

### Health checks

```bash
curl -sf http://localhost:8080/health/ready
GRPC_ADDR=localhost:50051 go run ./scripts/check_ready
```

## Environment files

| File | Commit? | Purpose |
|------|---------|---------|
| `.env.example` | Yes | Template for local dev |
| `.env.staging.example` | Yes | Template for staging deploy server |
| `.env.production.example` | Yes | Template for production deploy server |
| `.env` | **No** | Local overrides (gitignored) |
| `.env.staging` / `.env.production` | **No** | Server secrets (gitignored) |

Production deployments must inject env via the orchestrator — not via `.env` on disk.

## Migrations

| File | Phase | Content |
|------|-------|---------|
| `00001_baseline.sql` | F0 | `finance` schema |
| `00002_ledger.sql` | F1 | `ledger_journals`, `ledger_entries` |
| `00003_wallet_accounts.sql` | F2 | `wallet_accounts` (no balance column) |
| `00004_escrows.sql` | F3 | `escrows` lifecycle metadata |
| `00005_payment_orders.sql` | F4 | `payment_orders` |
| `00006_provider_events.sql` | F4 | `provider_events` audit |
| `00007_withdrawals.sql` | F6 | `withdrawals` (destination_hash only) |

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
