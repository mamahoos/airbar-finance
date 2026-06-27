# F8 Idempotency — Task Report

**Date:** 2026-06-26  
**Phase:** F8 — Idempotency (cross-cutting)  
**Branch:** feat/f8-idempotency  
**Depends on:** F0 Postgres/Redis, F3+ mutating gRPC handlers

---

## Objective

Deduplicate mutating finance commands with Postgres `idempotency_records`, Redis hot cache, and gRPC unary middleware per Scenario B docs.

---

## Deliverables

| ID | Task | Status |
|----|------|--------|
| F8.1 | Migration `00008_idempotency_records.sql` | Done |
| F8.2 | Idempotency middleware on all mutating RPCs | Done |
| F8.3 | Redis cache `idempotency:{key}` TTL 24h | Done |
| F8.4 | gRPC error mapping: Validation, Conflict, NotFound | Done |

---

## Schema

**`finance.idempotency_records`**

- `idempotency_key` PK
- `scope`, `resource_type`, `resource_id`
- `status`: PROCESSING | COMPLETED
- `response_snapshot` JSONB for replay

---

## Middleware flow

```text
mutating RPC
  → extract key (metadata idempotency-key OR RequestContext.idempotency_key)
  → Redis GET idempotency:{key}
  → miss → Postgres TryBeginProcessing
  → replay snapshot OR handler → Complete + Redis SET (24h)
  → handler error → DeleteProcessing rollback
```

**Mutating RPCs covered:** all Escrow commands except GetEscrow; Payment create/verify/topup; Withdrawal create/process/reject.

**Skipped (reads):** CheckReady, GetEscrow, GetPaymentOrder, GetWallet, ListWalletTransactions, ListWithdrawals.

---

## gRPC error mapping (F8.4)

| Domain error | gRPC code |
|--------------|-----------|
| `ErrKeyRequired` | `InvalidArgument` |
| `ErrConflict` (in-flight duplicate) | `Aborted` |
| `ErrNotFound` | `NotFound` |

---

## Gate F8

- [x] Duplicate key returns same response snapshot
- [x] Concurrent duplicate key while PROCESSING → conflict
- [x] Redis cache populated on complete (24h TTL)
- [x] Handler failure rolls back PROCESSING row

---

## Verification (2026-06-26)

```text
go vet ./...                                          — OK
go test ./... -count=1                                — OK
go test -tags=integration .../repository/ -count=1    — OK
go build ./cmd/server                                 — OK
```

---

## Next phase

**F9 — Audit:** `finance_events`, provider event audit trail.
