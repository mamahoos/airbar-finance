# Server Discovery — airbar VPS

> Snapshot from `ssh airbar` (2026-07-04). Facts for staging deploy design.

## Host

| Item | Value |
|------|-------|
| Hostname | `vps-f40babef` |
| CPU | 6 cores |
| RAM | 11 GiB (~7 GiB used) |
| Disk | 99 GiB (~81% used) |
| App root | `/srv/airbar.app/` |

## Shared infrastructure (`airbar-infra`)

| Service | Container | Network |
|---------|-----------|---------|
| PostgreSQL | `airbar-postgres` | `airbar-net` |
| Redis | `airbar-redis` | `airbar-net` |
| MinIO | `airbar-minio` | `airbar-net` |
| Nginx + SSL | `airbar-nginx` | `airbar-net` (80/443 public) |

Docker network: **`airbar-net`** (external, shared by all services).

## Legacy stack (do not replace yet)

Running under `/srv/airbar.app/`: `airbar-api`, `airbar-ffinance`, `airbar-front`, kafka, rabbitmq, notification, telegram, etc.

Existing public subdomains (nginx): `api.airbar.app`, `ffinance.airbar.app`, `admin.airbar.app`, ...

## New stack target (staging)

| Repo | Server path | Container | DB |
|------|-------------|-----------|-----|
| airbar-finance | `/srv/airbar.app/airbar-finance/` | `airbar-finance-app-staging` | `airbar_finance_staging` on `airbar-postgres` |
| airbar-core | `/srv/airbar.app/airbar-core/` | `airbar-core-app-staging` | `airbar_api_staging` on `airbar-postgres` |

Finance: internal gRPC only (`airbar-finance-app-staging:50051` on `airbar-net`).

Core: public HTTP via nginx → **`staging.api.airbar.app`** (DNS record by CTO).

## One-time server bootstrap

```bash
# On server — create staging databases (adjust user/password with infra team)
docker exec -it airbar-postgres psql -U postgres -c \
  "CREATE DATABASE airbar_finance_staging OWNER airbar;"
docker exec -it airbar-postgres psql -U postgres -c \
  "CREATE DATABASE airbar_api_staging OWNER airbar;"

# Create deploy dirs
sudo mkdir -p /srv/airbar.app/airbar-finance/migrations
sudo mkdir -p /srv/airbar.app/airbar-core
sudo chown -R debian:debian /srv/airbar.app/airbar-finance /srv/airbar.app/airbar-core

# Copy env templates (edit secrets on server)
cp .env.staging.example .env.staging   # in each repo dir
```

## Deploy flow

1. Merge to `main` → **CI** runs (tests, lint, build)
2. CI success → **Staging** workflow builds & pushes `:staging` + `:sha-<commit>` to GHCR (no SSH)
3. When infra is ready → run **Deploy — Staging** manually (`workflow_dispatch`) to SSH, migrate, and start containers
4. Health check via `docker exec` + wget inside container

Server deploy stays manual until staging bootstrap and secrets are explicitly approved.

See [staging-nginx-snippet.conf](./staging-nginx-snippet.conf) for CTO DNS/nginx request (`staging.api.airbar.app`).
