package audit_test

import (
	"context"
	"testing"

	domainaudit "github.com/mamahoos/airbar-finance/internal/domain/audit"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
)

type memoryFinanceEventRepo struct {
	events []*domainaudit.Event
}

func (m *memoryFinanceEventRepo) Create(_ context.Context, event *domainaudit.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *memoryFinanceEventRepo) CountByAggregate(_ context.Context, aggregateType, aggregateID string) (int64, error) {
	var count int64
	for _, e := range m.events {
		if e.AggregateType == aggregateType && e.AggregateID == aggregateID {
			count++
		}
	}
	return count, nil
}

func TestEmitterEmitEscrowCreated(t *testing.T) {
	repo := &memoryFinanceEventRepo{}
	emitter := audituc.NewEmitter(repo)

	if err := emitter.EmitEscrowCreated(context.Background(), "esc-1", "sh-1", "CREATED"); err != nil {
		t.Fatalf("EmitEscrowCreated() error = %v", err)
	}
	if len(repo.events) != 1 {
		t.Fatalf("events = %d, want 1", len(repo.events))
	}
	if repo.events[0].EventType != domainaudit.EventEscrowCreated {
		t.Fatalf("event type = %q", repo.events[0].EventType)
	}
}

func TestEmitterNilSafe(t *testing.T) {
	var emitter *audituc.Emitter
	if err := emitter.EmitEscrowCreated(context.Background(), "esc-1", "sh-1", "CREATED"); err != nil {
		t.Fatalf("nil emitter should no-op, got %v", err)
	}
}
