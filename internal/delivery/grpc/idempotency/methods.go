package idempotency

import (
	"fmt"
	"strings"

	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	"google.golang.org/protobuf/proto"
)

const (
	metadataKeyHeader = "idempotency-key"
)

type contextCarrier interface {
	GetContext() *financev1.RequestContext
}

type methodSpec struct {
	scope        string
	resourceType string
	newResponse  func() proto.Message
	resourceID   func(any) string
}

var mutatingMethods = map[string]methodSpec{
	"/airbar.finance.v1.EscrowService/CreateEscrow": {
		scope:        "escrow.create",
		resourceType: "shipment",
		newResponse:  func() proto.Message { return &financev1.EscrowResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.CreateEscrowRequest).GetShipmentId() },
	},
	"/airbar.finance.v1.EscrowService/FundEscrow": {
		scope:        "escrow.fund",
		resourceType: "shipment",
		newResponse:  func() proto.Message { return &financev1.EscrowResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.FundEscrowRequest).GetShipmentId() },
	},
	"/airbar.finance.v1.EscrowService/PayFromWallet": {
		scope:        "escrow.pay_from_wallet",
		resourceType: "shipment",
		newResponse:  func() proto.Message { return &financev1.EscrowResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.PayFromWalletRequest).GetShipmentId() },
	},
	"/airbar.finance.v1.EscrowService/MarkDelivered": {
		scope:        "escrow.mark_delivered",
		resourceType: "shipment",
		newResponse:  func() proto.Message { return &financev1.EscrowResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.MarkDeliveredRequest).GetShipmentId() },
	},
	"/airbar.finance.v1.EscrowService/FreezeEscrow": {
		scope:        "escrow.freeze",
		resourceType: "shipment",
		newResponse:  func() proto.Message { return &financev1.EscrowResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.FreezeEscrowRequest).GetShipmentId() },
	},
	"/airbar.finance.v1.EscrowService/ReleaseEscrow": {
		scope:        "escrow.release",
		resourceType: "shipment",
		newResponse:  func() proto.Message { return &financev1.EscrowResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.ReleaseEscrowRequest).GetShipmentId() },
	},
	"/airbar.finance.v1.EscrowService/RefundEscrow": {
		scope:        "escrow.refund",
		resourceType: "shipment",
		newResponse:  func() proto.Message { return &financev1.EscrowResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.RefundEscrowRequest).GetShipmentId() },
	},
	"/airbar.finance.v1.EscrowService/PartialRefundEscrow": {
		scope:        "escrow.partial_refund",
		resourceType: "shipment",
		newResponse:  func() proto.Message { return &financev1.EscrowResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.PartialRefundEscrowRequest).GetShipmentId() },
	},
	"/airbar.finance.v1.PaymentOrderService/CreatePaymentOrder": {
		scope:        "payment.create",
		resourceType: "shipment",
		newResponse:  func() proto.Message { return &financev1.PaymentOrderResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.CreatePaymentOrderRequest).GetShipmentId() },
	},
	"/airbar.finance.v1.PaymentOrderService/VerifyPaymentOrder": {
		scope:        "payment.verify",
		resourceType: "payment_order",
		newResponse:  func() proto.Message { return &financev1.PaymentOrderResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.VerifyPaymentOrderRequest).GetOrderId() },
	},
	"/airbar.finance.v1.PaymentOrderService/CreateWalletTopupOrder": {
		scope:        "wallet.topup.create",
		resourceType: "user",
		newResponse:  func() proto.Message { return &financev1.PaymentOrderResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.CreateWalletTopupRequest).GetUserId() },
	},
	"/airbar.finance.v1.PaymentOrderService/VerifyWalletTopupOrder": {
		scope:        "wallet.topup.verify",
		resourceType: "payment_order",
		newResponse:  func() proto.Message { return &financev1.PaymentOrderResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.VerifyPaymentOrderRequest).GetOrderId() },
	},
	"/airbar.finance.v1.WithdrawalService/CreateWithdrawal": {
		scope:        "withdrawal.create",
		resourceType: "user",
		newResponse:  func() proto.Message { return &financev1.WithdrawalResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.CreateWithdrawalRequest).GetUserId() },
	},
	"/airbar.finance.v1.WithdrawalService/ProcessWithdrawal": {
		scope:        "withdrawal.process",
		resourceType: "withdrawal",
		newResponse:  func() proto.Message { return &financev1.WithdrawalResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.ProcessWithdrawalRequest).GetWithdrawalId() },
	},
	"/airbar.finance.v1.WithdrawalService/RejectWithdrawal": {
		scope:        "withdrawal.reject",
		resourceType: "withdrawal",
		newResponse:  func() proto.Message { return &financev1.WithdrawalResponse{} },
		resourceID:   func(req any) string { return req.(*financev1.RejectWithdrawalRequest).GetWithdrawalId() },
	},
}

func lookupMethod(fullMethod string) (methodSpec, bool) {
	spec, ok := mutatingMethods[fullMethod]
	return spec, ok
}

func extractIdempotencyKey(metadataKey, bodyKey string) string {
	if key := strings.TrimSpace(metadataKey); key != "" {
		return key
	}
	return strings.TrimSpace(bodyKey)
}

func bodyIdempotencyKey(req any) string {
	carrier, ok := req.(contextCarrier)
	if !ok || carrier.GetContext() == nil {
		return ""
	}
	return carrier.GetContext().GetIdempotencyKey()
}

func snapshotMethodKey(fullMethod string) string {
	return fmt.Sprintf("@method:%s", fullMethod)
}
