package escrow

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	credituc "github.com/mamahoos/airbar-finance/internal/usecase/credit"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
)

// RefundEscrowInput is the application input for UC-07.
type RefundEscrowInput struct {
	ShipmentID string
}

// RefundEscrow credits the full remaining escrow balance to payer promo credit and/or wallet.
type RefundEscrow struct {
	pool         *pgxpool.Pool
	escrowRepo   domainescrow.Repository
	postJournal  *ledgeruc.PostJournal
	ledger       LedgerBalanceReader
	ensureCredit *credituc.EnsureCreditAccount
	audit        *audituc.Emitter
}

// NewRefundEscrow creates the RefundEscrow use case.
func NewRefundEscrow(
	pool *pgxpool.Pool,
	escrowRepo domainescrow.Repository,
	postJournal *ledgeruc.PostJournal,
	ledger LedgerBalanceReader,
	ensureCredit *credituc.EnsureCreditAccount,
	audit *audituc.Emitter,
) *RefundEscrow {
	return &RefundEscrow{
		pool:         pool,
		escrowRepo:   escrowRepo,
		postJournal:  postJournal,
		ledger:       ledger,
		ensureCredit: ensureCredit,
		audit:        audit,
	}
}

// Execute posts refund journals and transitions escrow to REFUNDED.
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

		promoRefund, walletRefund := RefundAllocation(balance, escrow.PromoCreditFunded)
		if err := ensureCreditAccountIfNeeded(txCtx, uc.ensureCredit, escrow.PayerUserID, promoRefund); err != nil {
			return err
		}
		if err := postEscrowRefundJournals(txCtx, uc.postJournal, escrow.PayerUserID, input.ShipmentID, promoRefund, walletRefund); err != nil {
			return err
		}

		now := nowUTC()
		escrow.PromoCreditFunded -= promoRefund
		if escrow.PromoCreditFunded < 0 {
			escrow.PromoCreditFunded = 0
		}
		escrow.Status = domainescrow.StatusRefunded
		escrow.RefundedAt = &now
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

func ensureCreditAccountIfNeeded(
	ctx context.Context,
	ensureCredit *credituc.EnsureCreditAccount,
	payerUserID string,
	promoRefund int64,
) error {
	if promoRefund <= 0 || ensureCredit == nil {
		return nil
	}
	_, err := ensureCredit.Execute(ctx, payerUserID)
	return err
}

func postEscrowRefundJournals(
	ctx context.Context,
	postJournal *ledgeruc.PostJournal,
	payerUserID, shipmentID string,
	promoRefund, walletRefund int64,
) error {
	for _, input := range BuildEscrowRefundJournalLines(payerUserID, shipmentID, promoRefund, walletRefund) {
		if _, err := postJournal.Execute(ctx, input); err != nil {
			return err
		}
	}
	return nil
}
