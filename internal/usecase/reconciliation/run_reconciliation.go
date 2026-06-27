package reconciliation

import (
	"context"
	"time"

	domainrecon "github.com/mamahoos/airbar-finance/internal/domain/reconciliation"
)

// GlobalLedgerReader reads global ledger sums for reconciliation.
type GlobalLedgerReader interface {
	SumGlobal(ctx context.Context) (debit int64, credit int64, err error)
}

// RunReconciliation checks global debit=credit and stores a run (UC-21).
type RunReconciliation struct {
	ledger GlobalLedgerReader
	runs   domainrecon.Repository
}

// NewRunReconciliation creates the RunReconciliation use case.
func NewRunReconciliation(ledger GlobalLedgerReader, runs domainrecon.Repository) *RunReconciliation {
	return &RunReconciliation{ledger: ledger, runs: runs}
}

// Execute runs the global ledger check and persists the outcome.
func (uc *RunReconciliation) Execute(ctx context.Context) (*domainrecon.Run, error) {
	started := time.Now().UTC()

	totalDebit, totalCredit, err := uc.ledger.SumGlobal(ctx)
	if err != nil {
		return nil, err
	}

	imbalance := totalDebit - totalCredit
	balanced := imbalance == 0

	status := domainrecon.StatusPassed
	if !balanced {
		status = domainrecon.StatusFailed
	}

	findings := map[string]any{
		"global_debit":  totalDebit,
		"global_credit": totalCredit,
		"imbalance":     imbalance,
		"balanced":      balanced,
	}

	completed := time.Now().UTC()
	run := &domainrecon.Run{
		Status:      status,
		Findings:    findings,
		StartedAt:   started,
		CompletedAt: &completed,
	}

	if err := uc.runs.Create(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}
