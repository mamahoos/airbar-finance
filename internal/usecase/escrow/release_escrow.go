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

// ReleaseEscrowInput is the application input for UC-06.
type ReleaseEscrowInput struct {
	ShipmentID string
}

// ReleaseEscrow splits escrow to platform fee and carrier wallet.
type ReleaseEscrow struct {
	pool               *pgxpool.Pool
	escrowRepo         domainescrow.Repository
	postJournal        *ledgeruc.PostJournal
	ledger             LedgerBalanceReader
	platformFeePercent float64
}

// NewReleaseEscrow creates the ReleaseEscrow use case.
func NewReleaseEscrow(
	pool *pgxpool.Pool,
	escrowRepo domainescrow.Repository,
	postJournal *ledgeruc.PostJournal,
	ledger LedgerBalanceReader,
	platformFeePercent float64,
) *ReleaseEscrow {
	return &ReleaseEscrow{
		pool:               pool,
		escrowRepo:         escrowRepo,
		postJournal:        postJournal,
		ledger:             ledger,
		platformFeePercent: platformFeePercent,
	}
}

// Execute posts ESCROW_RELEASE and transitions escrow to RELEASED.
func (uc *ReleaseEscrow) Execute(ctx context.Context, input ReleaseEscrowInput) (*domainescrow.Escrow, error) {
	if input.ShipmentID == "" {
		return nil, domainescrow.ErrInvalidAmount
	}

	var result *domainescrow.Escrow
	err := pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		escrow, err := uc.escrowRepo.GetByShipmentID(txCtx, input.ShipmentID)
		if err != nil {
			return err
		}
		if !escrow.Status.CanRelease() {
			return domainescrow.ErrInvalidTransition
		}

		balance, err := EscrowBalance(txCtx, uc.ledger, input.ShipmentID)
		if err != nil {
			return err
		}
		if balance <= 0 {
			return domainescrow.ErrNoEscrowBalance
		}

		fee := domainescrow.CalcPlatformFee(balance, uc.platformFeePercent)
		carrierAmount := balance - fee

		_, err = uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypeEscrowRelease,
			RefID:       fmt.Sprintf("%s:release", input.ShipmentID),
			Description: "Release escrow to carrier",
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.ShipmentEscrowAccount(input.ShipmentID), Debit: balance, Credit: 0},
				{AccountCode: domainledger.AccountAirbarFeeRevenue, Debit: 0, Credit: fee},
				{AccountCode: domainledger.UserWalletAccount(escrow.CarrierUserID), Debit: 0, Credit: carrierAmount},
			},
		})
		if err != nil {
			return err
		}

		now := nowUTC()
		escrow.Status = domainescrow.StatusReleased
		escrow.ReleasedAt = &now
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
