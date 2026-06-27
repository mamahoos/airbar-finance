package escrow

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
)

// RefundEscrowInput is the application input for UC-07.
type RefundEscrowInput struct {
	ShipmentID string
}

// RefundEscrow credits the full remaining escrow balance to the payer wallet.
type RefundEscrow struct {
	pool        *pgxpool.Pool
	escrowRepo  domainescrow.Repository
	postJournal *ledgeruc.PostJournal
	ledger      LedgerBalanceReader
}

// NewRefundEscrow creates the RefundEscrow use case.
func NewRefundEscrow(
	pool *pgxpool.Pool,
	escrowRepo domainescrow.Repository,
	postJournal *ledgeruc.PostJournal,
	ledger LedgerBalanceReader,
) *RefundEscrow {
	return &RefundEscrow{
		pool:        pool,
		escrowRepo:  escrowRepo,
		postJournal: postJournal,
		ledger:      ledger,
	}
}

// Execute posts ESCROW_REFUND_WALLET and transitions escrow to REFUNDED.
func (uc *RefundEscrow) Execute(ctx context.Context, input RefundEscrowInput) (*domainescrow.Escrow, error) {
	if input.ShipmentID == "" {
		return nil, domainescrow.ErrInvalidAmount
	}

	var result *domainescrow.Escrow
	err := pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		escrow, err := uc.escrowRepo.GetByShipmentID(txCtx, input.ShipmentID)
		if err != nil {
			return err
		}
		if !escrow.Status.CanRefund() {
			return domainescrow.ErrInvalidTransition
		}

		balance, err := EscrowBalance(txCtx, uc.ledger, input.ShipmentID)
		if err != nil {
			return err
		}
		if balance <= 0 {
			return domainescrow.ErrNoEscrowBalance
		}

		_, err = uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypeEscrowRefundWallet,
			RefID:       fmt.Sprintf("%s:refund", input.ShipmentID),
			Description: "Refund escrow to payer wallet",
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.ShipmentEscrowAccount(input.ShipmentID), Debit: balance, Credit: 0},
				{AccountCode: domainledger.UserWalletAccount(escrow.PayerUserID), Debit: 0, Credit: balance},
			},
		})
		if err != nil {
			return err
		}

		now := nowUTC()
		escrow.Status = domainescrow.StatusRefunded
		escrow.RefundedAt = &now
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
