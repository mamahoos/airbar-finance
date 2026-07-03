package escrow

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	credituc "github.com/mamahoos/airbar-finance/internal/usecase/credit"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
)

// PayFromWalletInput is the application input for UC-03.
type PayFromWalletInput struct {
	ShipmentID  string
	PayerUserID string
	Amount      int64
}

// PayFromWallet debits payer promo credit first, then wallet, and funds escrow.
type PayFromWallet struct {
	pool          *pgxpool.Pool
	escrowRepo    domainescrow.Repository
	postJournal   *ledgeruc.PostJournal
	getWallet     *walletuc.GetBalance
	getPromo      *credituc.GetBalance
	ensureCredit  *credituc.EnsureCreditAccount
	audit         *audituc.Emitter
}

// NewPayFromWallet creates the PayFromWallet use case.
func NewPayFromWallet(
	pool *pgxpool.Pool,
	escrowRepo domainescrow.Repository,
	postJournal *ledgeruc.PostJournal,
	getWallet *walletuc.GetBalance,
	getPromo *credituc.GetBalance,
	ensureCredit *credituc.EnsureCreditAccount,
	audit *audituc.Emitter,
) *PayFromWallet {
	return &PayFromWallet{
		pool:         pool,
		escrowRepo:   escrowRepo,
		postJournal:  postJournal,
		getWallet:    getWallet,
		getPromo:     getPromo,
		ensureCredit: ensureCredit,
		audit:        audit,
	}
}

// Execute posts promo-first payer funding and transitions escrow to FUNDED.
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

		promoBalance, err := uc.getPromo.Execute(txCtx, input.PayerUserID)
		if err != nil {
			return err
		}
		walletBalance, err := uc.getWallet.Execute(txCtx, input.PayerUserID)
		if err != nil {
			return err
		}

		split, err := ResolvePayerFundingSplit(input.Amount, promoBalance, walletBalance)
		if err != nil {
			return err
		}
		if split.PromoCredit > 0 {
			if _, err := uc.ensureCredit.Execute(txCtx, input.PayerUserID); err != nil {
				return err
			}
		}

		_, err = uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
			RefType:     RefTypeForEscrowPay(split),
			RefID:       fmt.Sprintf("%s:wallet-pay", input.ShipmentID),
			Description: "Pay escrow from promo credit and wallet",
			Lines:       BuildEscrowPayJournalLines(input.PayerUserID, input.ShipmentID, split),
		})
		if err != nil {
			return err
		}

		now := nowUTC()
		escrow.Status = domainescrow.StatusFunded
		escrow.FundingSource = FundingSourceForSplit(split)
		escrow.PromoCreditFunded = split.PromoCredit
		escrow.FundedAt = &now
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
