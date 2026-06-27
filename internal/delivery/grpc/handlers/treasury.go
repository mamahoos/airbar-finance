package handlers

import (
	"context"
	"errors"

	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	treasuryuc "github.com/mamahoos/airbar-finance/internal/usecase/treasury"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// TreasuryHandler implements TreasuryService (UC-20).
type TreasuryHandler struct {
	financev1.UnimplementedTreasuryServiceServer
	getTreasurySummary *treasuryuc.GetTreasurySummary
}

// NewTreasuryHandler creates a TreasuryService gRPC handler.
func NewTreasuryHandler(getTreasurySummary *treasuryuc.GetTreasurySummary) *TreasuryHandler {
	return &TreasuryHandler{getTreasurySummary: getTreasurySummary}
}

func (h *TreasuryHandler) GetTreasurySummary(ctx context.Context, req *financev1.GetTreasuryRequest) (*financev1.TreasurySummaryResponse, error) {
	summary, err := h.getTreasurySummary.Execute(ctx, req.GetCurrency())
	if err != nil {
		return nil, mapTreasuryError(err)
	}

	accounts := make(map[string]any, len(summary.Accounts))
	for code, balance := range summary.Accounts {
		accounts[code] = walletuc.FormatAmount(balance)
	}
	structAccounts, err := structpb.NewStruct(accounts)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &financev1.TreasurySummaryResponse{
		Currency: summary.Currency,
		Accounts: structAccounts,
	}, nil
}

func mapTreasuryError(err error) error {
	if errors.Is(err, treasuryuc.ErrUnsupportedCurrency) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, "internal error")
}
