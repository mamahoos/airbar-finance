# F6 Withdrawal — Task Report

**Date:** 2026-06-26  
**Phase:** F6 — Withdrawal (UC-16..19)  
**Branch:** feat/f6-withdrawal  
**Depends on:** F2 Wallet Accounts, F1 Ledger

---

## Objective

Carrier payout requests with wallet reserve, admin process/reject, and PII-safe destination storage.

---

## Deliverables

| ID | Task | Status |
|----|------|--------|
| F6.1 | Migration `00007_withdrawals.sql` | Done |
| F6.2 | `CreateWithdrawal` + `WITHDRAWAL_RESERVE` (UC-16) | Done |
| F6.3 | `ListWithdrawals` (UC-17) | Done |
| F6.4 | `ProcessWithdrawal` → COMPLETED (UC-18) | Done |
| F6.5 | `RejectWithdrawal` + `WITHDRAWAL_REJECT_REVERSAL` (UC-19) | Done |
| F6.6 | gRPC `WithdrawalService` handlers | Done |

---

## Schema

**`finance.withdrawals`**

- `destination_hash` only — plain IBAN never stored
- `status`: PENDING | COMPLETED | REJECTED

---

## Ledger journals

**WITHDRAWAL_RESERVE (CreateWithdrawal)**

```text
Debit:  USER:{id}:WALLET_LIABILITY   amount
Credit: IR_PAYOUT_CLEARING           amount
```

**WITHDRAWAL_REJECT_REVERSAL (RejectWithdrawal)**

```text
Debit:  IR_PAYOUT_CLEARING           amount
Credit: USER:{id}:WALLET_LIABILITY   amount
```

---

## Gate F6

- [x] Reserve debits wallet and credits payout clearing
- [x] Reject reverses reserve to wallet
- [x] Process marks COMPLETED without double-spend
- [x] IBAN hashed at boundary — not persisted plain

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

**F7 — Treasury + Reconciliation:** UC-20..23.
