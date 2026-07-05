# Infra patch — gray-cloud for nested staging DNS

Apply in `/srv/airbar.app/airbar-infra/scripts/cloudflare/sync.sh` on the VPS.
There is no separate `airbar-infra` git repo yet; keep this file in sync with server state.

## Why

Cloudflare Universal SSL covers `*.airbar.app` but **not** `*.*.airbar.app`.
Records like `staging.api.airbar.app` must stay **DNS-only** (gray cloud) or browsers get
`ERR_SSL_VERSION_OR_CIPHER_MISMATCH`.

## Patch

Add to `PROXY_OFF_DOMAINS` (keep existing entries):

```bash
PROXY_OFF_DOMAINS=(
  "freellm.${ZONE}"
  "staging.api.${ZONE}"
  "staging.finance.${ZONE}"
  "staging.app.${ZONE}"
)
```

## Apply on server

```bash
cd /srv/airbar.app/airbar-infra
# edit sync.sh as above, then:
./scripts/cloudflare/sync.sh dns
./scripts/cloudflare/sync.sh proxy-off   # idempotent if already gray
```

Verify:

```bash
dig +short staging.api.airbar.app @1.1.1.1
# must return origin IP (158.69.206.33), not 172.67.x.x
```
