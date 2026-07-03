package handlers

import (
	"context"

	domainwithdrawal "github.com/mamahoos/airbar-finance/internal/domain/withdrawal"
	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	withdrawaluc "github.com/mamahoos/airbar-finance/internal/usecase/withdrawal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WithdrawalHandler implements WithdrawalService (UC-16..19).
type WithdrawalHandler struct {
	financev1.UnimplementedWithdrawalServiceServer
	createWithdrawal  *withdrawaluc.CreateWithdrawal
	listWithdrawals   *withdrawaluc.ListWithdrawals
	processWithdrawal *withdrawaluc.ProcessWithdrawal
	rejectWithdrawal  *withdrawaluc.RejectWithdrawal
}

// NewWithdrawalHandler creates a WithdrawalService gRPC handler.
func NewWithdrawalHandler(
	createWithdrawal *withdrawaluc.CreateWithdrawal,
	listWithdrawals *withdrawaluc.ListWithdrawals,
	processWithdrawal *withdrawaluc.ProcessWithdrawal,
	rejectWithdrawal *withdrawaluc.RejectWithdrawal,
) *WithdrawalHandler {
	return &WithdrawalHandler{
		createWithdrawal:  createWithdrawal,
		listWithdrawals:   listWithdrawals,
		processWithdrawal: processWithdrawal,
		rejectWithdrawal:  rejectWithdrawal,
	}
}

func (h *WithdrawalHandler) CreateWithdrawal(ctx context.Context, req *financev1.CreateWithdrawalRequest) (*financev1.WithdrawalResponse, error) {
	amount, err := withdrawaluc.ParseAmount(req.GetAmount())
	if err != nil {
		return nil, mapWithdrawalError(err)
	}

	withdrawal, err := h.createWithdrawal.Execute(ctx, withdrawaluc.CreateWithdrawalInput{
		UserID:               req.GetUserId(),
		Amount:               amount,
		DestinationIBAN:      req.GetDestinationIban(),
		UserActive:           req.GetUserActive(),
		FinancialKycApproved: req.GetFinancialKycApproved(),
	})
	if err != nil {
		return nil, mapWithdrawalError(err)
	}
	return toWithdrawalResponse(withdrawal), nil
}

func (h *WithdrawalHandler) ListWithdrawals(ctx context.Context, req *financev1.ListWithdrawalsRequest) (*financev1.WithdrawalsResponse, error) {
	items, err := h.listWithdrawals.Execute(ctx, req.GetUserId(), req.GetStatus())
	if err != nil {
		return nil, mapWithdrawalError(err)
	}

	resp := &financev1.WithdrawalsResponse{
		Items: make([]*financev1.WithdrawalResponse, len(items)),
	}
	for i := range items {
		resp.Items[i] = toWithdrawalResponse(&items[i])
	}
	return resp, nil
}

func (h *WithdrawalHandler) ProcessWithdrawal(ctx context.Context, req *financev1.ProcessWithdrawalRequest) (*financev1.WithdrawalResponse, error) {
	withdrawal, err := h.processWithdrawal.Execute(ctx, withdrawaluc.ProcessWithdrawalInput{
		WithdrawalID:  req.GetWithdrawalId(),
		ProviderRef:   req.GetProviderRef(),
		PayoutChannel: req.GetPayoutChannel(),
		ReceiptURL:    req.GetReceiptUrl(),
	})
	if err != nil {
		return nil, mapWithdrawalError(err)
	}
	return toWithdrawalResponse(withdrawal), nil
}

func (h *WithdrawalHandler) RejectWithdrawal(ctx context.Context, req *financev1.RejectWithdrawalRequest) (*financev1.WithdrawalResponse, error) {
	withdrawal, err := h.rejectWithdrawal.Execute(ctx, withdrawaluc.RejectWithdrawalInput{
		WithdrawalID: req.GetWithdrawalId(),
		Reason:       req.GetReason(),
	})
	if err != nil {
		return nil, mapWithdrawalError(err)
	}
	return toWithdrawalResponse(withdrawal), nil
}

func toWithdrawalResponse(withdrawal *domainwithdrawal.Withdrawal) *financev1.WithdrawalResponse {
	if withdrawal == nil {
		return nil
	}
	resp := &financev1.WithdrawalResponse{
		Id:            withdrawal.ID,
		UserId:        withdrawal.UserID,
		Amount:        withdrawaluc.FormatAmount(withdrawal.Amount),
		Status:        string(withdrawal.Status),
		ProviderRef:   withdrawal.ProviderRef,
		PayoutChannel: withdrawal.PayoutChannel,
		ReceiptUrl:    withdrawal.ReceiptURL,
		CreatedAt:     timestamppb.New(withdrawal.CreatedAt),
	}
	if withdrawal.ProcessedAt != nil {
		resp.ProcessedAt = timestamppb.New(*withdrawal.ProcessedAt)
	}
	return resp
}
