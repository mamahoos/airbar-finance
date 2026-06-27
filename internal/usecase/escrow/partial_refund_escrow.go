package escrow

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
)

// PartialRefundEscrowInput is the application input for UC-08.
type PartialRefundEscrowInput struct {
	ShipmentID   string
	RefundAmount int64
}

// PartialRefundEscrow credits part of escrow to the payer wallet.
type PartialRefundEscrow struct {
	pool        *pgxpool.Pool
	escrowRepo  domainescrow.Repository
	postJournal *ledgeruc.PostJournal
	ledger      LedgerBalanceReader
	audit       *audituc.Emitter
}

// NewPartialRefundEscrow creates the PartialRefundEscrow use case.
func NewPartialRefundEscrow(
	pool *pgxpool.Pool,
	escrowRepo domainescrow.Repository,
	postJournal *ledgeruc.PostJournal,
	ledger LedgerBalanceReader,
	audit *audituc.Emitter,
) *PartialRefundEscrow {
	return &PartialRefundEscrow{
		pool:        pool,
		escrowRepo:  escrowRepo,
		postJournal: postJournal,
		ledger:      ledger,
		audit:       audit,
	}
}

// Execute posts a partial ESCROW_REFUND_WALLET journal.
func (uc *PartialRefundEscrow) Execute(ctx context.Context, input PartialRefundEscrowInput) (*domainescrow.Escrow, error) {
	if input.ShipmentID == "" || input.RefundAmount <= 0 {
		return nil, domainescrow.ErrInvalidAmount
	}

	var result *domainescrow.Escrow
	err := pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		escrow, err := uc.escrowRepo.GetByShipmentID(txCtx, input.ShipmentID)
		if err != nil {
			return err
		}
		if !escrow.Status.CanPartialRefund() {
			return domainescrow.ErrInvalidTransition
		}

		balance, err := EscrowBalance(txCtx, uc.ledger, input.ShipmentID)
		if err != nil {
			return err
		}
		if balance <= 0 {
			return domainescrow.ErrNoEscrowBalance
		}
		if input.RefundAmount > balance {
			return domainescrow.ErrRefundExceedsBalance
		}

		_, err = uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypeEscrowRefundWallet,
			RefID:       fmt.Sprintf("%s:partial-refund:%d", input.ShipmentID, input.RefundAmount),
			Description: "Partial refund escrow to payer wallet",
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.ShipmentEscrowAccount(input.ShipmentID), Debit: input.RefundAmount, Credit: 0},
				{AccountCode: domainledger.UserWalletAccount(escrow.PayerUserID), Debit: 0, Credit: input.RefundAmount},
			},
		})
		if err != nil {
			return err
		}

		now := nowUTC()
		if input.RefundAmount == balance {
			escrow.Status = domainescrow.StatusRefunded
			escrow.RefundedAt = &now
		} else {
			escrow.Status = domainescrow.StatusPartiallyRefunded
			escrow.RefundedAt = &now
		}
		if err := uc.escrowRepo.Update(txCtx, escrow); err != nil {
			return err
		}
		emitEscrowStatus(txCtx, uc.audit, escrow)
		result = escrow
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
