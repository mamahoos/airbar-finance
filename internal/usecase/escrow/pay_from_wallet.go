package escrow

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
)

// PayFromWalletInput is the application input for UC-03.
type PayFromWalletInput struct {
	ShipmentID  string
	PayerUserID string
	Amount      int64
}

// PayFromWallet debits payer wallet and funds escrow from WALLET source.
type PayFromWallet struct {
	pool        *pgxpool.Pool
	escrowRepo  domainescrow.Repository
	postJournal *ledgeruc.PostJournal
	getBalance  *walletuc.GetBalance
}

// NewPayFromWallet creates the PayFromWallet use case.
func NewPayFromWallet(
	pool *pgxpool.Pool,
	escrowRepo domainescrow.Repository,
	postJournal *ledgeruc.PostJournal,
	getBalance *walletuc.GetBalance,
) *PayFromWallet {
	return &PayFromWallet{
		pool:        pool,
		escrowRepo:  escrowRepo,
		postJournal: postJournal,
		getBalance:  getBalance,
	}
}

// Execute posts WALLET_TO_ESCROW and transitions escrow to FUNDED.
func (uc *PayFromWallet) Execute(ctx context.Context, input PayFromWalletInput) (*domainescrow.Escrow, error) {
	if input.ShipmentID == "" || input.PayerUserID == "" || input.Amount <= 0 {
		return nil, domainescrow.ErrInvalidAmount
	}

	var result *domainescrow.Escrow
	err := pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		escrow, err := uc.escrowRepo.GetByShipmentID(txCtx, input.ShipmentID)
		if err != nil {
			return err
		}
		if escrow.PayerUserID != input.PayerUserID {
			return domainescrow.ErrPayerMismatch
		}
		if escrow.Amount != input.Amount {
			return domainescrow.ErrAmountMismatch
		}
		if !escrow.Status.CanPayFromWallet() {
			return domainescrow.ErrInvalidTransition
		}

		balance, err := uc.getBalance.Execute(txCtx, input.PayerUserID)
		if err != nil {
			return err
		}
		if balance < input.Amount {
			return domainescrow.ErrInsufficientWallet
		}

		_, err = uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypeWalletToEscrow,
			RefID:       fmt.Sprintf("%s:wallet-pay", input.ShipmentID),
			Description: "Pay escrow from wallet",
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.UserWalletAccount(input.PayerUserID), Debit: input.Amount, Credit: 0},
				{AccountCode: domainledger.ShipmentEscrowAccount(input.ShipmentID), Debit: 0, Credit: input.Amount},
			},
		})
		if err != nil {
			return err
		}

		now := nowUTC()
		escrow.Status = domainescrow.StatusFunded
		escrow.FundingSource = domainescrow.FundingSourceWallet
		escrow.FundedAt = &now
		if err := uc.escrowRepo.Update(txCtx, escrow); err != nil {
			return err
		}
		result = escrow
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
