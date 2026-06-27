# F5 Wallet Queries — Task Report

**Date:** 2026-06-26  
**Phase:** F5 — Wallet Queries (UC-14, UC-15)  
**Branch:** feat/f5-wallet-queries  
**Depends on:** F2 Wallet Accounts, F1 Ledger

---

## Objective

Expose read-only wallet APIs for Node to proxy balance and transaction history from ledger SSOT.

---

## Deliverables

| ID | Task | Status |
|----|------|--------|
| F5.1 | `GetWallet` — balance + account metadata (UC-14) | Done |
| F5.2 | `ListWalletTransactions` from journals (UC-15) | Done |
| F5.3 | gRPC `WalletService` handlers | Done |

---

## Gate F5

- [x] Balance derived from `SUM(ledger_entries)` — no stored balance
- [x] Transaction history joined from `ledger_journals` + `ledger_entries`
- [x] Node can proxy via `WalletService.GetWallet` / `ListWalletTransactions`

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

**F6 — Withdrawal:** migration `withdrawals`, UC-16..19, gRPC WithdrawalService.
