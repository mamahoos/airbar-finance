# F9 Audit — Task Report

**Date:** 2026-06-26  
**Phase:** F9 — Audit  
**Branch:** feat/f9-audit  
**Depends on:** F3 Escrow, F4 Payment, F6 Withdrawal

---

## Objective

Immutable audit trail for finance aggregate state changes plus verified Zibal provider event coverage.

---

## Deliverables

| ID | Task | Status |
|----|------|--------|
| F9.1 | Migration `00010_finance_events.sql` | Done |
| F9.2 | Emit on escrow/payment/withdrawal state change | Done |
| F9.3 | `provider_events` on Zibal request/verify/callback | Verified + integration assertion |

---

## Schema

**`finance.finance_events`**

- `aggregate_type`: escrow | payment_order | withdrawal
- `aggregate_id`, `event_type`, `payload` JSONB

---

## Emitters

`usecase/audit.Emitter` wired into escrow, payment, and withdrawal use cases inside the same DB transaction where status changes.

---

## Gate F9

- [x] Escrow create writes `finance_events`
- [x] Payment direct flow has REQUEST + VERIFY `provider_events`
- [x] No proto contract changes

---

## Verification (2026-06-26)

```text
go vet ./...
go test ./... -count=1
go test -tags=integration .../repository/ -run Integration -count=1
go build ./cmd/server
```

---

## v1 finance complete

F6 + F7 + F8 + F9 on branch stack. **F10 Hardening** remains a separate follow-up PR.
