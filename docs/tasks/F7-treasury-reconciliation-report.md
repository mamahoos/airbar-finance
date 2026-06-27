# F7 Treasury + Reconciliation — Task Report

**Date:** 2026-06-27  
**Phase:** F7 — Treasury + Reconciliation (UC-20..23)  
**Branch:** feat/f7-treasury-reconciliation  
**Depends on:** F1 Ledger Core

---

## Objective

Ops visibility into system account balances and automated ledger integrity checks with persisted reconciliation runs.

---

## Deliverables

| ID | Task | Status |
|----|------|--------|
| F7.1 | Migration `00008_reconciliation_runs.sql` | Done |
| F7.2 | `GetTreasurySummary` (UC-20) | Done |
| F7.3 | `RunReconciliation` — global debit=credit (UC-21) | Done |
| F7.4 | `ListReconciliationRuns`, `GetReconciliationRun` (UC-22, UC-23) | Done |
| F7.5 | gRPC `TreasuryService` + `ReconciliationService` | Done |

---

## Treasury summary accounts

| Key | Meaning |
|-----|---------|
| `IR_PSP_CLEARING` | Zibal clearing (asset net) |
| `IR_BANK_MAIN` | Bank treasury (asset net) |
| `IR_PAYOUT_CLEARING` | Payout queue (asset net) |
| `AIRBAR_FEE_REVENUE` | Platform fee revenue |
| `AGGREGATE_WALLET_LIABILITY` | Sum of user wallet liabilities |
| `AGGREGATE_ESCROW_LIABILITY` | Sum of shipment escrow liabilities |

---

## Gate F7

- [x] GetTreasurySummary returns system + aggregate balances from ledger SSOT
- [x] RunReconciliation PASSED when global debit=credit
- [x] Runs persisted and listable via gRPC
- [x] Integration test: topup → treasury → recon → escrow shift

---

## Next phase

**F8 — Idempotency:** cross-cutting dedup on mutating RPCs.
