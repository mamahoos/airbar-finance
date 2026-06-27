# F3 Escrow — Task Report

**Date:** 2026-06-26  
**Phase:** F3 — Escrow (UC-01..08)  
**Branch:** feat/f3-escrow  
**Depends on:** F2 Wallet Accounts

---

## Objective

Implement shipment escrow lifecycle with state machine guards, ledger-backed funding/release/refund, and gRPC `EscrowService`.

---

## Deliverables

| ID | Task | Status |
|----|------|--------|
| F3.1 | Migration `00004_escrows.sql` | Done |
| F3.2 | Domain status enum + transition guards | Done |
| F3.3 | `CreateEscrow` (UC-01) | Done |
| F3.4 | `GetEscrow` (UC-02) | Done |
| F3.5 | `FundEscrow` (internal / gRPC) | Done |
| F3.6 | `PayFromWallet` + `WALLET_TO_ESCROW` (UC-03) | Done |
| F3.7 | `MarkDelivered` → `DISPUTE_WINDOW` (UC-04) | Done |
| F3.8 | `FreezeEscrow` (UC-05) | Done |
| F3.9 | `ReleaseEscrow` — fee + carrier wallet (UC-06) | Done |
| F3.10 | `RefundEscrow` — payer wallet parity (UC-07) | Done |
| F3.11 | `PartialRefundEscrow` (UC-08) | Done |
| F3.12 | gRPC `EscrowService` handlers | Done |
| F3.13 | State machine integration tests | Done |

---

## Schema

**`finance.escrows`**

- `shipment_id` (unique), `carrier_user_id`, `payer_user_id`, `amount`
- `status`, `payment_order_id`, `funding_source` (`PSP` | `WALLET`)
- lifecycle timestamps: `funded_at`, `released_at`, `refunded_at`
- **no balance column** — escrow balance from ledger `SHIPMENT:{id}:IRT:ESCROW`

---

## State machine

```text
CREATED → FUNDED → DISPUTE_WINDOW → RELEASED | REFUNDED | PARTIALLY_REFUNDED
              ↓              ↓
    PayFromWallet/Fund   FROZEN → REFUNDED only (ReleaseEscrow blocked)
```

---

## Gate F3

- [x] WALLET pay end-to-end: CreateEscrow → PayFromWallet → MarkDelivered → ReleaseEscrow
- [x] Release split: platform fee + carrier wallet credit
- [x] Refund always credits payer wallet (wallet parity)
- [x] FROZEN blocks ReleaseEscrow; RefundEscrow succeeds

---

## Verification (2026-06-26)

```text
go vet ./...                                          — OK
go test ./... -count=1                                — OK
go test -tags=integration .../repository/ -count=1  — OK
go build ./cmd/server                                 — OK
```

---

## Next phase

**F4 — Payment + Zibal:** `payment_orders`, Zibal adapter, UC-09..13.
