package ledger

import (
	"context"
	"errors"
	"testing"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

type mockRepo struct {
	createFn func(ctx context.Context, journal *domainledger.Journal) error
	sumFn    func(ctx context.Context, accountCode domainledger.AccountCode) (int64, int64, error)
}

func (m *mockRepo) CreateJournal(ctx context.Context, journal *domainledger.Journal) error {
	if m.createFn != nil {
		return m.createFn(ctx, journal)
	}
	journal.ID = "journal-1"
	return nil
}

func (m *mockRepo) SumByAccount(ctx context.Context, accountCode domainledger.AccountCode) (int64, int64, error) {
	if m.sumFn != nil {
		return m.sumFn(ctx, accountCode)
	}
	return 0, 0, nil
}

func (m *mockRepo) SumGlobal(_ context.Context) (int64, int64, error) {
	return 0, 0, nil
}

func (m *mockRepo) SumByAccountLike(_ context.Context, _ string) (int64, int64, error) {
	return 0, 0, nil
}

func (m *mockRepo) ListByAccount(_ context.Context, _ domainledger.AccountCode) ([]domainledger.AccountEntry, error) {
	return nil, nil
}

func TestPostJournalSuccess(t *testing.T) {
	uc := NewPostJournal(&mockRepo{}, nil)

	journal, err := uc.Execute(context.Background(), PostJournalInput{
		RefType: domainledger.RefTypeWalletToEscrow,
		RefID:   "sh-1:pay",
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.UserWalletAccount("payer"), Debit: 5000, Credit: 0},
			{AccountCode: domainledger.ShipmentEscrowAccount("sh-1"), Debit: 0, Credit: 5000},
		},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if journal.ID != "journal-1" {
		t.Fatalf("journal.ID = %q, want journal-1", journal.ID)
	}
}

func TestPostJournalUnbalanced(t *testing.T) {
	uc := NewPostJournal(&mockRepo{}, nil)

	_, err := uc.Execute(context.Background(), PostJournalInput{
		RefType: domainledger.RefTypeWalletToEscrow,
		RefID:   "sh-2:pay",
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.UserWalletAccount("payer"), Debit: 100, Credit: 0},
			{AccountCode: domainledger.ShipmentEscrowAccount("sh-2"), Debit: 0, Credit: 50},
		},
	})
	if !errors.Is(err, domainledger.ErrUnbalancedJournal) {
		t.Fatalf("Execute() error = %v, want ErrUnbalancedJournal", err)
	}
}

func TestPostJournalDuplicate(t *testing.T) {
	uc := NewPostJournal(&mockRepo{
		createFn: func(_ context.Context, _ *domainledger.Journal) error {
			return domainledger.ErrDuplicateJournal
		},
	}, nil)

	_, err := uc.Execute(context.Background(), PostJournalInput{
		RefType: domainledger.RefTypeWalletToEscrow,
		RefID:   "sh-3:pay",
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.UserWalletAccount("payer"), Debit: 100, Credit: 0},
			{AccountCode: domainledger.ShipmentEscrowAccount("sh-3"), Debit: 0, Credit: 100},
		},
	})
	if !errors.Is(err, domainledger.ErrDuplicateJournal) {
		t.Fatalf("Execute() error = %v, want ErrDuplicateJournal", err)
	}
}
