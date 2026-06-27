package reconciliation

import (
	"context"
	"testing"

	domainrecon "github.com/mamahoos/airbar-finance/internal/domain/reconciliation"
)

type stubLedger struct {
	debit  int64
	credit int64
}

func (s *stubLedger) SumGlobal(_ context.Context) (int64, int64, error) {
	return s.debit, s.credit, nil
}

type stubRuns struct {
	created []*domainrecon.Run
}

func (s *stubRuns) Create(_ context.Context, run *domainrecon.Run) error {
	run.ID = "run-1"
	s.created = append(s.created, run)
	return nil
}

func (s *stubRuns) GetByID(_ context.Context, id string) (*domainrecon.Run, error) {
	for _, run := range s.created {
		if run.ID == id {
			return run, nil
		}
	}
	return nil, domainrecon.ErrNotFound
}

func (s *stubRuns) List(_ context.Context) ([]domainrecon.Run, error) {
	out := make([]domainrecon.Run, len(s.created))
	for i, run := range s.created {
		out[i] = *run
	}
	return out, nil
}

func TestRunReconciliationBalanced(t *testing.T) {
	runs := &stubRuns{}
	uc := NewRunReconciliation(&stubLedger{debit: 1000, credit: 1000}, runs)

	run, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if run.Status != domainrecon.StatusPassed {
		t.Fatalf("status = %q, want PASSED", run.Status)
	}
	if run.Findings["balanced"] != true {
		t.Fatalf("findings balanced = %v", run.Findings["balanced"])
	}
}

func TestRunReconciliationFailed(t *testing.T) {
	runs := &stubRuns{}
	uc := NewRunReconciliation(&stubLedger{debit: 1000, credit: 900}, runs)

	run, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if run.Status != domainrecon.StatusFailed {
		t.Fatalf("status = %q, want FAILED", run.Status)
	}
	if run.Findings["imbalance"] != int64(100) {
		t.Fatalf("imbalance = %v", run.Findings["imbalance"])
	}
}
