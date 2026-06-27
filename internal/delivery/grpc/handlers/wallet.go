package handlers

import (
	"context"

	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WalletHandler implements WalletService (UC-14..15).
type WalletHandler struct {
	financev1.UnimplementedWalletServiceServer
	getWallet              *walletuc.GetWallet
	listWalletTransactions *walletuc.ListWalletTransactions
}

// NewWalletHandler creates a WalletService gRPC handler.
func NewWalletHandler(
	getWallet *walletuc.GetWallet,
	listWalletTransactions *walletuc.ListWalletTransactions,
) *WalletHandler {
	return &WalletHandler{
		getWallet:              getWallet,
		listWalletTransactions: listWalletTransactions,
	}
}

func (h *WalletHandler) GetWallet(ctx context.Context, req *financev1.GetWalletRequest) (*financev1.WalletResponse, error) {
	wallet, err := h.getWallet.Execute(ctx, req.GetUserId(), req.GetCurrency())
	if err != nil {
		return nil, mapWalletError(err)
	}
	return &financev1.WalletResponse{
		UserId:      wallet.UserID,
		Currency:    wallet.Currency,
		Balance:     walletuc.FormatAmount(wallet.Balance),
		AccountCode: wallet.AccountCode,
	}, nil
}

func (h *WalletHandler) ListWalletTransactions(ctx context.Context, req *financev1.ListWalletTransactionsRequest) (*financev1.WalletTransactionsResponse, error) {
	items, err := h.listWalletTransactions.Execute(ctx, req.GetUserId(), req.GetCurrency())
	if err != nil {
		return nil, mapWalletError(err)
	}

	resp := &financev1.WalletTransactionsResponse{
		Items: make([]*financev1.WalletTransactionItem, len(items)),
	}
	for i, item := range items {
		resp.Items[i] = &financev1.WalletTransactionItem{
			JournalId:   item.JournalID,
			RefType:     item.RefType,
			RefId:       item.RefID,
			Description: item.Description,
			Debit:       walletuc.FormatAmount(item.Debit),
			Credit:      walletuc.FormatAmount(item.Credit),
			CreatedAt:   timestamppb.New(item.CreatedAt),
		}
	}
	return resp, nil
}
