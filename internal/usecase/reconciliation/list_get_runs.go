package reconciliation

import (
	"context"

	domainrecon "github.com/mamahoos/airbar-finance/internal/domain/reconciliation"
)

// ListReconciliationRuns returns reconciliation history (UC-22).
type ListReconciliationRuns struct {
	runs domainrecon.Repository
}

// NewListReconciliationRuns creates the ListReconciliationRuns use case.
func NewListReconciliationRuns(runs domainrecon.Repository) *ListReconciliationRuns {
	return &ListReconciliationRuns{runs: runs}
}

// Execute lists runs ordered by started_at descending.
func (uc *ListReconciliationRuns) Execute(ctx context.Context) ([]domainrecon.Run, error) {
	return uc.runs.List(ctx)
}

// GetReconciliationRun loads a run by id (UC-23).
type GetReconciliationRun struct {
	runs domainrecon.Repository
}

// NewGetReconciliationRun creates the GetReconciliationRun use case.
func NewGetReconciliationRun(runs domainrecon.Repository) *GetReconciliationRun {
	return &GetReconciliationRun{runs: runs}
}

// Execute returns a single reconciliation run.
func (uc *GetReconciliationRun) Execute(ctx context.Context, runID string) (*domainrecon.Run, error) {
	if runID == "" {
		return nil, domainrecon.ErrNotFound
	}
	return uc.runs.GetByID(ctx, runID)
}
