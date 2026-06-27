# F4 Payment + Zibal — Task Report

**Date:** 2026-06-26  
**Phase:** F4 — Payment + Zibal (UC-09..13)  
**Branch:** feat/f4-zibal  
**Depends on:** F3 Escrow

---

## Objective

Zibal DIRECT shipment pay and wallet topup: payment orders, provider audit, HTTP callback, verify → fund escrow or credit wallet.

---

## Deliverables

| ID | Task | Status |
|----|------|--------|
| F4.1 | Migration `00005_payment_orders.sql` | Done |
| F4.2 | Migration `00006_provider_events.sql` | Done |
| F4.3 | Zibal client Request + Verify | Done |
| F4.4 | `CreatePaymentOrder` SHIPMENT (UC-09) | Done |
| F4.5 | `GetPaymentOrder` (UC-10) | Done |
| F4.6 | HTTP `GET /api/v1/zibal/callback` | Done |
| F4.7 | `VerifyPaymentOrder` → FundEscrow (UC-11) | Done |
| F4.8 | `CreateWalletTopupOrder` (UC-12) | Done |
| F4.9 | `VerifyWalletTopupOrder` + `WALLET_TOPUP` (UC-13) | Done |
| F4.10 | gRPC `PaymentOrderService` handlers | Done |
| F4.11 | Integration tests with mock Zibal | Done |

---

## Gate F4

- [x] DIRECT pay: CreateEscrow → CreatePaymentOrder → Verify → escrow FUNDED (PSP)
- [x] Wallet topup: CreateWalletTopupOrder → Verify → wallet credited
- [x] Callback route on finance HTTP edge
- [x] `provider_events` logged on request/verify/callback

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

**F5 — Wallet queries:** GetWallet, ListWalletTransactions, gRPC WalletService.
