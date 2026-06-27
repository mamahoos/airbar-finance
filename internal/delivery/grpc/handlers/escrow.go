package handlers

import (
	"context"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	escrowuc "github.com/mamahoos/airbar-finance/internal/usecase/escrow"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EscrowHandler implements EscrowService (UC-01..08).
type EscrowHandler struct {
	financev1.UnimplementedEscrowServiceServer
	createEscrow        *escrowuc.CreateEscrow
	getEscrow           *escrowuc.GetEscrow
	fundEscrow          *escrowuc.FundEscrow
	payFromWallet       *escrowuc.PayFromWallet
	markDelivered       *escrowuc.MarkDelivered
	freezeEscrow        *escrowuc.FreezeEscrow
	releaseEscrow       *escrowuc.ReleaseEscrow
	refundEscrow        *escrowuc.RefundEscrow
	partialRefundEscrow *escrowuc.PartialRefundEscrow
}

// NewEscrowHandler creates an EscrowService gRPC handler.
func NewEscrowHandler(
	createEscrow *escrowuc.CreateEscrow,
	getEscrow *escrowuc.GetEscrow,
	fundEscrow *escrowuc.FundEscrow,
	payFromWallet *escrowuc.PayFromWallet,
	markDelivered *escrowuc.MarkDelivered,
	freezeEscrow *escrowuc.FreezeEscrow,
	releaseEscrow *escrowuc.ReleaseEscrow,
	refundEscrow *escrowuc.RefundEscrow,
	partialRefundEscrow *escrowuc.PartialRefundEscrow,
) *EscrowHandler {
	return &EscrowHandler{
		createEscrow:        createEscrow,
		getEscrow:           getEscrow,
		fundEscrow:          fundEscrow,
		payFromWallet:       payFromWallet,
		markDelivered:       markDelivered,
		freezeEscrow:        freezeEscrow,
		releaseEscrow:       releaseEscrow,
		refundEscrow:        refundEscrow,
		partialRefundEscrow: partialRefundEscrow,
	}
}

func (h *EscrowHandler) CreateEscrow(ctx context.Context, req *financev1.CreateEscrowRequest) (*financev1.EscrowResponse, error) {
	amount, err := escrowuc.ParseAmount(req.GetAmount())
	if err != nil {
		return nil, mapEscrowError(err)
	}

	escrow, err := h.createEscrow.Execute(ctx, escrowuc.CreateEscrowInput{
		ShipmentID:    req.GetShipmentId(),
		CarrierUserID: req.GetCarrierUserId(),
		PayerUserID:   req.GetPayerUserId(),
		Amount:        amount,
	})
	if err != nil {
		return nil, mapEscrowError(err)
	}
	return toEscrowResponse(escrow), nil
}

func (h *EscrowHandler) GetEscrow(ctx context.Context, req *financev1.GetEscrowRequest) (*financev1.EscrowResponse, error) {
	escrow, err := h.getEscrow.Execute(ctx, req.GetShipmentId())
	if err != nil {
		return nil, mapEscrowError(err)
	}
	return toEscrowResponse(escrow), nil
}

func (h *EscrowHandler) FundEscrow(ctx context.Context, req *financev1.FundEscrowRequest) (*financev1.EscrowResponse, error) {
	escrow, err := h.fundEscrow.Execute(ctx, escrowuc.FundEscrowInput{
		ShipmentID:     req.GetShipmentId(),
		PaymentOrderID: req.GetPaymentOrderId(),
	})
	if err != nil {
		return nil, mapEscrowError(err)
	}
	return toEscrowResponse(escrow), nil
}

func (h *EscrowHandler) PayFromWallet(ctx context.Context, req *financev1.PayFromWalletRequest) (*financev1.EscrowResponse, error) {
	amount, err := escrowuc.ParseAmount(req.GetAmount())
	if err != nil {
		return nil, mapEscrowError(err)
	}

	escrow, err := h.payFromWallet.Execute(ctx, escrowuc.PayFromWalletInput{
		ShipmentID:  req.GetShipmentId(),
		PayerUserID: req.GetPayerUserId(),
		Amount:      amount,
	})
	if err != nil {
		return nil, mapEscrowError(err)
	}
	return toEscrowResponse(escrow), nil
}

func (h *EscrowHandler) MarkDelivered(ctx context.Context, req *financev1.MarkDeliveredRequest) (*financev1.EscrowResponse, error) {
	escrow, err := h.markDelivered.Execute(ctx, escrowuc.MarkDeliveredInput{
		ShipmentID: req.GetShipmentId(),
	})
	if err != nil {
		return nil, mapEscrowError(err)
	}
	return toEscrowResponse(escrow), nil
}

func (h *EscrowHandler) FreezeEscrow(ctx context.Context, req *financev1.FreezeEscrowRequest) (*financev1.EscrowResponse, error) {
	escrow, err := h.freezeEscrow.Execute(ctx, escrowuc.FreezeEscrowInput{
		ShipmentID: req.GetShipmentId(),
	})
	if err != nil {
		return nil, mapEscrowError(err)
	}
	return toEscrowResponse(escrow), nil
}

func (h *EscrowHandler) ReleaseEscrow(ctx context.Context, req *financev1.ReleaseEscrowRequest) (*financev1.EscrowResponse, error) {
	escrow, err := h.releaseEscrow.Execute(ctx, escrowuc.ReleaseEscrowInput{
		ShipmentID: req.GetShipmentId(),
	})
	if err != nil {
		return nil, mapEscrowError(err)
	}
	return toEscrowResponse(escrow), nil
}

func (h *EscrowHandler) RefundEscrow(ctx context.Context, req *financev1.RefundEscrowRequest) (*financev1.EscrowResponse, error) {
	escrow, err := h.refundEscrow.Execute(ctx, escrowuc.RefundEscrowInput{
		ShipmentID: req.GetShipmentId(),
	})
	if err != nil {
		return nil, mapEscrowError(err)
	}
	return toEscrowResponse(escrow), nil
}

func (h *EscrowHandler) PartialRefundEscrow(ctx context.Context, req *financev1.PartialRefundEscrowRequest) (*financev1.EscrowResponse, error) {
	refundAmount, err := escrowuc.ParseAmount(req.GetRefundAmount())
	if err != nil {
		return nil, mapEscrowError(err)
	}

	escrow, err := h.partialRefundEscrow.Execute(ctx, escrowuc.PartialRefundEscrowInput{
		ShipmentID:   req.GetShipmentId(),
		RefundAmount: refundAmount,
	})
	if err != nil {
		return nil, mapEscrowError(err)
	}
	return toEscrowResponse(escrow), nil
}

func toEscrowResponse(escrow *domainescrow.Escrow) *financev1.EscrowResponse {
	if escrow == nil {
		return nil
	}
	resp := &financev1.EscrowResponse{
		Id:             escrow.ID,
		ShipmentId:     escrow.ShipmentID,
		CarrierUserId:  escrow.CarrierUserID,
		PayerUserId:    escrow.PayerUserID,
		Amount:         escrowuc.FormatAmount(escrow.Amount),
		Status:         string(escrow.Status),
		PaymentOrderId: escrow.PaymentOrderID,
		FundingSource:  string(escrow.FundingSource),
		CreatedAt:      timestamppb.New(escrow.CreatedAt),
		UpdatedAt:      timestamppb.New(escrow.UpdatedAt),
	}
	if escrow.FundedAt != nil {
		resp.FundedAt = timestamppb.New(*escrow.FundedAt)
	}
	if escrow.ReleasedAt != nil {
		resp.ReleasedAt = timestamppb.New(*escrow.ReleasedAt)
	}
	if escrow.RefundedAt != nil {
		resp.RefundedAt = timestamppb.New(*escrow.RefundedAt)
	}
	return resp
}
