package handlers

import (
	"context"

	domainpayment "github.com/mamahoos/airbar-finance/internal/domain/payment"
	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	paymentuc "github.com/mamahoos/airbar-finance/internal/usecase/payment"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PaymentHandler implements PaymentOrderService (UC-09..13).
type PaymentHandler struct {
	financev1.UnimplementedPaymentOrderServiceServer
	createPaymentOrder     *paymentuc.CreatePaymentOrder
	getPaymentOrder        *paymentuc.GetPaymentOrder
	verifyPaymentOrder     *paymentuc.VerifyPaymentOrder
	createWalletTopupOrder *paymentuc.CreateWalletTopupOrder
	verifyWalletTopupOrder *paymentuc.VerifyWalletTopupOrder
}

// NewPaymentHandler creates a PaymentOrderService gRPC handler.
func NewPaymentHandler(
	createPaymentOrder *paymentuc.CreatePaymentOrder,
	getPaymentOrder *paymentuc.GetPaymentOrder,
	verifyPaymentOrder *paymentuc.VerifyPaymentOrder,
	createWalletTopupOrder *paymentuc.CreateWalletTopupOrder,
	verifyWalletTopupOrder *paymentuc.VerifyWalletTopupOrder,
) *PaymentHandler {
	return &PaymentHandler{
		createPaymentOrder:     createPaymentOrder,
		getPaymentOrder:        getPaymentOrder,
		verifyPaymentOrder:     verifyPaymentOrder,
		createWalletTopupOrder: createWalletTopupOrder,
		verifyWalletTopupOrder: verifyWalletTopupOrder,
	}
}

func (h *PaymentHandler) CreatePaymentOrder(ctx context.Context, req *financev1.CreatePaymentOrderRequest) (*financev1.PaymentOrderResponse, error) {
	amount, err := paymentuc.ParseAmount(req.GetAmount())
	if err != nil {
		return nil, mapPaymentError(err)
	}
	var agreedPrice int64
	if req.GetAgreedPrice() != "" {
		agreedPrice, err = paymentuc.ParseAmount(req.GetAgreedPrice())
		if err != nil {
			return nil, mapPaymentError(err)
		}
	}

	order, err := h.createPaymentOrder.Execute(ctx, paymentuc.CreatePaymentOrderInput{
		ShipmentID:  req.GetShipmentId(),
		PayerUserID: req.GetPayerUserId(),
		Amount:      amount,
		SuccessURL:  req.GetSuccessUrl(),
		FailureURL:  req.GetFailureUrl(),
		Description: req.GetDescription(),
		AgreedPrice: agreedPrice,
	})
	if err != nil {
		return nil, mapPaymentError(err)
	}
	return toPaymentOrderResponse(order), nil
}

func (h *PaymentHandler) GetPaymentOrder(ctx context.Context, req *financev1.GetPaymentOrderRequest) (*financev1.PaymentOrderResponse, error) {
	order, err := h.getPaymentOrder.Execute(ctx, req.GetOrderId())
	if err != nil {
		return nil, mapPaymentError(err)
	}
	return toPaymentOrderResponse(order), nil
}

func (h *PaymentHandler) VerifyPaymentOrder(ctx context.Context, req *financev1.VerifyPaymentOrderRequest) (*financev1.PaymentOrderResponse, error) {
	order, err := h.verifyPaymentOrder.Execute(ctx, req.GetOrderId(), req.GetAuthority())
	if err != nil {
		return nil, mapPaymentError(err)
	}
	return toPaymentOrderResponse(order), nil
}

func (h *PaymentHandler) CreateWalletTopupOrder(ctx context.Context, req *financev1.CreateWalletTopupRequest) (*financev1.PaymentOrderResponse, error) {
	amount, err := paymentuc.ParseAmount(req.GetAmount())
	if err != nil {
		return nil, mapPaymentError(err)
	}

	order, err := h.createWalletTopupOrder.Execute(ctx, paymentuc.CreateWalletTopupOrderInput{
		UserID:      req.GetUserId(),
		Amount:      amount,
		SuccessURL:  req.GetSuccessUrl(),
		FailureURL:  req.GetFailureUrl(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, mapPaymentError(err)
	}
	return toPaymentOrderResponse(order), nil
}

func (h *PaymentHandler) VerifyWalletTopupOrder(ctx context.Context, req *financev1.VerifyPaymentOrderRequest) (*financev1.PaymentOrderResponse, error) {
	order, err := h.verifyWalletTopupOrder.Execute(ctx, req.GetOrderId(), req.GetAuthority())
	if err != nil {
		return nil, mapPaymentError(err)
	}
	return toPaymentOrderResponse(order), nil
}

func toPaymentOrderResponse(order *domainpayment.Order) *financev1.PaymentOrderResponse {
	if order == nil {
		return nil
	}
	resp := &financev1.PaymentOrderResponse{
		Id:          order.ID,
		ShipmentId:  order.ShipmentID,
		PayerUserId: order.PayerUserID,
		Purpose:     string(order.Purpose),
		Amount:      paymentuc.FormatAmount(order.Amount),
		Status:      string(order.Status),
		Authority:   order.Authority,
		RedirectUrl: order.RedirectURL,
		CreatedAt:   timestamppb.New(order.CreatedAt),
	}
	if order.VerifiedAt != nil {
		resp.VerifiedAt = timestamppb.New(*order.VerifiedAt)
	}
	return resp
}
