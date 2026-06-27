# F2 Wallet Accounts — Task Report

**Date:** 2026-06-26  
**Phase:** F2 — Wallet Accounts  
**Branch:** feat/f2-wallet  
**Depends on:** F1 Ledger Core

---

## Objective

Register user wallet accounts without storing balance. Balance is derived from ledger SSOT:

`balance = SUM(credit) - SUM(debit)` on `USER:{id}:IRT:WALLET_LIABILITY`

---

## Deliverables

| ID | Task | Status |
|----|------|--------|
| F2.1 | Migration `00003_wallet_accounts.sql` | Done |
| F2.2 | `EnsureWalletAccount` lazy create | Done |
| F2.3 | `GetBalance(user_id)` from ledger sums | Done |

---

## Schema

**`finance.wallet_accounts`**

- `user_id`, `currency` (IRT), `account_code` — **no balance column**
- Unique `(user_id, currency)` and `account_code`

---

## Implementation

| Layer | Path |
|-------|------|
| Domain | `internal/domain/wallet/` |
| Use cases | `EnsureWalletAccount`, `GetBalance`, `EnsureForLines` |
| Repository | `internal/infrastructure/postgres/repository/wallet.go` |
| Ledger touch | `PostJournal` calls `EnsureForLines` before persist |

---

## Gate F2

- [x] No stored balance column
- [x] Lazy create on first ledger touch (via PostJournal)
- [x] GetBalance from ledger entries only

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

**F3 — Escrow:** migration `escrows`, state machine, UC-01..08.
