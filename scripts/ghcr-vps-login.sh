#!/usr/bin/env bash
# Persist GHCR credentials on the VPS for manual `docker pull`.
# Requires a classic PAT with read:packages (store as GitHub secret GHCR_READ_TOKEN).
#
# Usage (on VPS as root):
#   GHCR_READ_TOKEN='ghp_...' ./scripts/ghcr-vps-login.sh
# Or read token from env file:
#   set -a; source /srv/airbar.app/.secrets/ghcr.env; set +a
#   ./scripts/ghcr-vps-login.sh
set -euo pipefail

GHCR_USER="${GHCR_USER:-mamahoos}"
TOKEN="${GHCR_READ_TOKEN:-${1:-}}"

if [[ -z "$TOKEN" ]]; then
  echo "Usage: GHCR_READ_TOKEN=ghp_... $0" >&2
  echo "Create PAT: GitHub → Settings → Developer settings → PAT (classic) → read:packages" >&2
  exit 1
fi

echo "$TOKEN" | docker login ghcr.io -u "$GHCR_USER" --password-stdin
echo "GHCR login OK for ${GHCR_USER} (credentials in ~/.docker/config.json)"
