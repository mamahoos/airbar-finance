# F0 Bootstrap — Task Report

**Date:** 2026-06-26  
**Phase:** F0 — Bootstrap  
**PR:** feat/f0-bootstrap  
**UC covered:** UC-24 (`FinanceHealthService.CheckReady`)

---

## Objective

Deliver a deployable `airbar-finance` process with:

- Environment-based configuration
- Postgres + Redis connectivity checks
- gRPC `CheckReady` (UC-24)
- HTTP `GET /health/ready`
- Structured logging and graceful shutdown
- CI: `go vet`, `go test`, `go build`
- Docker image build

No ledger, escrow, or payment business logic in this phase.

---

## What was implemented

### Application bootstrap

| Component | Path | Description |
|-----------|------|-------------|
| Main | `cmd/server/main.go` | Starts gRPC + HTTP servers; graceful shutdown on SIGINT/SIGTERM |
| Config | `internal/infrastructure/config/` | Loads **all** required env vars; optional `.env` via godotenv (local only) |
| Postgres | `internal/infrastructure/postgres/` | `pgxpool` connection factory |
| Redis | `internal/infrastructure/redis/` | Client from URL + ping helper |
| Readiness | `internal/infrastructure/health/` | `Ready()` pings Postgres and Redis |
| gRPC | `internal/delivery/grpc/` | Server registration + `FinanceHealthService` handler |
| HTTP | `internal/delivery/http/` | `/health/ready` → 200 `ok` or 503 `not ready` |

### Proto / codegen

- Updated `proto/airbar_finance_v1.proto` `go_package` to match module path
- Generated stubs in `internal/gen/financev1/` (committed for CI without `protoc`)

### Tooling & infra

| File | Change |
|------|--------|
| `Makefile` | `proto`, `build`, `test`, `vet`, `verify` targets |
| `.github/workflows/ci.yml` | Enabled `go vet`, `go test`, `go build` |
| `Dockerfile` | Multi-stage build → `airbar-finance` binary |
| `docker-compose.yml` | Enabled `airbar-finance` service; Redis host port `6381` (avoids clash with other local Redis on `6379`) |
| `.env.example` | `REDIS_URL=redis://localhost:6381/1` |
| `scripts/check_ready/` | Dev helper to call gRPC `CheckReady` |

### Tests

| Package | Tests |
|---------|-------|
| `internal/infrastructure/config` | Required fields validation; full env parsing from `.env.example` keys |
| `internal/delivery/http` | Ready → 200; not ready → 503 |

---

## Verification results

### Static checks (Docker `golang:1.22-bookworm`)

```text
go vet ./...     — OK
go test ./...    — OK (unit tests)
go build ./cmd/server — OK
```

### Docker image

```text
docker build -t airbar-finance:local .  — OK
```

### Runtime (integration)

Stack: `postgres-finance` on host `5434`, Redis on host `6379` (existing local instance), app container `airbar-finance:local`.

| Check | Command / endpoint | Result |
|-------|-------------------|--------|
| HTTP ready | `curl http://localhost:8080/health/ready` | `200` body `ok` |
| gRPC ready | `go run ./scripts/check_ready` (`GRPC_ADDR=localhost:50051`) | `ready=true` |
| Container logs | JSON slog | gRPC + HTTP listening messages |

---

## Architecture notes

- **Dependency rule:** `delivery` → `infrastructure` ports; no domain logic yet
- **Readiness:** Both Postgres and Redis must respond to `Ping`; otherwise `ready=false` / HTTP 503
- **Config:** All settings from env / `.env` (local); no hardcoded defaults in `config.Load()`
- **Public HTTP surface:** Only `/health/ready` in F0 (Zibal callback in F4)
- **Build:** `-buildvcs=false` in Docker/CI when `.git` is unavailable or incomplete inside build context

---

## Out of scope (superseded by F1+)

- ~~F1 Ledger~~ — see [F1-ledger-report.md](./F1-ledger-report.md)
- F3 Escrow use cases
- F4 Zibal + payment orders
- F8 Idempotency middleware
- gRPC reflection (F10)

---

## Local run

```bash
cp .env.example .env
make up          # postgres-finance + redis (Redis on host 6381)
make migrate-up  # optional for F0 health (ping only needs DB up)
docker build -t airbar-finance:local .
docker compose up -d airbar-finance   # or run binary with .env
curl http://localhost:8080/health/ready
GRPC_ADDR=localhost:50051 go run ./scripts/check_ready
```

---

## Files added / changed (summary)

**New:** `cmd/server/main.go`, config, postgres, redis, health, grpc/http delivery, `internal/gen/financev1/*`, `scripts/check_ready/main.go`, `go.sum`, this report.

**Updated:** `go.mod`, `proto`, `Makefile`, `Dockerfile`, `docker-compose.yml`, `ci.yml`, `.env.example`.

**Removed:** obsolete `.gitkeep` files in populated directories.
