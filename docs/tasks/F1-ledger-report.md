# F1 Ledger Core — Task Report

**Date:** 2026-06-26  
**Phase:** F1 — Ledger Core  
**Branch:** feat/f1-ledger  
**Depends on:** F0 Bootstrap

---

## Objective

Implement the double-entry ledger as SSOT for money:

- Atomic `PostJournal` with `debit = credit` invariant
- System account code helpers
- Postgres persistence with duplicate journal guard
- Balance query via `SUM(ledger_entries)` per account

No gRPC exposure yet (escrow/payment use cases wire this in F3+).

---

## Deliverables

| ID | Task | Status |
|----|------|--------|
| F1.1 | Migration `00002_ledger.sql` | Done |
| F1.2 | Domain `Journal`, `Entry`, `Repository` port | Done |
| F1.3 | System account constants + helpers | Done |
| F1.4 | `PostJournal` use case + Postgres repository | Done |
| F1.5 | Unit + integration tests | Done |

---

## Schema

**`finance.ledger_journals`**

- `id`, `ref_type`, `ref_id`, `description`, `created_at`
- Unique `(ref_type, ref_id)` — duplicate business events rejected

**`finance.ledger_entries`**

- `id`, `journal_id`, `account_code`, `debit`, `credit`, `created_at`
- CHECK: exactly one side per line; amounts ≥ 0

---

## Account codes (F1.3)

| Code | Helper |
|------|--------|
| `IR_PSP_CLEARING` | constant |
| `IR_BANK_MAIN` | constant |
| `IR_PAYOUT_CLEARING` | constant |
| `AIRBAR_FEE_REVENUE` | constant |
| `USER:{id}:IRT:WALLET_LIABILITY` | `UserWalletAccount(userID)` |
| `SHIPMENT:{id}:IRT:ESCROW` | `ShipmentEscrowAccount(shipmentID)` |

---

## Gate F1

- [x] Every journal balanced before persist
- [x] Duplicate `(ref_type, ref_id)` → `ErrDuplicateJournal`
- [x] `SumByAccount` returns debit/credit totals from DB

---

## Verification (2026-06-26)

```text
go vet ./...                                          — OK
go test ./... -count=1                                — OK (all unit tests)
go test .../repository/ -run Integration -count=1   — OK
go build ./cmd/server                                 — OK
docker build -t airbar-finance:local .                — OK
```

---

## Tests

| Package | Coverage |
|---------|----------|
| `domain/ledger` | ValidateLines, account code format |
| `usecase/ledger` | PostJournal success, unbalanced, duplicate |
| `postgres/repository` | Integration (requires `TEST_DATABASE_URL`) |

```bash
# Unit tests (CI)
go vet ./...
go test ./...

# Integration (local, after migrate)
export TEST_DATABASE_URL=postgres://airbar:airbar@localhost:5434/airbar_finance?sslmode=disable
make migrate-up
go test ./internal/infrastructure/postgres/repository/ -run Integration -v
```

---

## Next phase

**F2 — Wallet accounts:** `wallet_accounts` migration, lazy create, `GetBalance` from ledger sums.
