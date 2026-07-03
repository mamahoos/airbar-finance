package withdrawal

import (
	"context"
	"testing"

	domainwithdrawal "github.com/mamahoos/airbar-finance/internal/domain/withdrawal"
)

type mockWithdrawalRepo struct {
	items map[string]*domainwithdrawal.Withdrawal
}

func (m *mockWithdrawalRepo) Create(_ context.Context, withdrawal *domainwithdrawal.Withdrawal) error {
	if m.items == nil {
		m.items = make(map[string]*domainwithdrawal.Withdrawal)
	}
	copy := *withdrawal
	copy.ID = "wd-1"
	m.items[copy.ID] = &copy
	*withdrawal = copy
	return nil
}

func (m *mockWithdrawalRepo) GetByID(_ context.Context, id string) (*domainwithdrawal.Withdrawal, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, domainwithdrawal.ErrNotFound
	}
	copy := *item
	return &copy, nil
}

func (m *mockWithdrawalRepo) List(_ context.Context, userID string, status domainwithdrawal.Status) ([]domainwithdrawal.Withdrawal, error) {
	var items []domainwithdrawal.Withdrawal
	for _, item := range m.items {
		if item.UserID != userID {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		items = append(items, *item)
	}
	return items, nil
}

func (m *mockWithdrawalRepo) Update(_ context.Context, withdrawal *domainwithdrawal.Withdrawal) error {
	m.items[withdrawal.ID] = withdrawal
	return nil
}

func TestProcessWithdrawalFromPending(t *testing.T) {
	repo := &mockWithdrawalRepo{
		items: map[string]*domainwithdrawal.Withdrawal{
			"wd-1": {ID: "wd-1", UserID: "user-1", Amount: 1000, Status: domainwithdrawal.StatusPending},
		},
	}
	uc := NewProcessWithdrawal(repo, nil)

	result, err := uc.Execute(context.Background(), ProcessWithdrawalInput{
		WithdrawalID:  "wd-1",
		ProviderRef:   "bank-ref-1",
		PayoutChannel: "PAYA",
		ReceiptURL:    "https://receipts.example/wd-1",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != domainwithdrawal.StatusCompleted {
		t.Fatalf("status = %q, want COMPLETED", result.Status)
	}
	if result.ProviderRef != "bank-ref-1" || result.PayoutChannel != "PAYA" || result.ReceiptURL == "" {
		t.Fatalf("receipt fields not persisted: %+v", result)
	}
}

func TestWithdrawalLifecycleApproveSentSettled(t *testing.T) {
	repo := &mockWithdrawalRepo{
		items: map[string]*domainwithdrawal.Withdrawal{
			"wd-1": {ID: "wd-1", UserID: "user-1", Amount: 1000, Status: domainwithdrawal.StatusPending},
		},
	}

	approved, err := NewApproveWithdrawal(repo, nil).Execute(context.Background(), "wd-1")
	if err != nil {
		t.Fatalf("ApproveWithdrawal() error = %v", err)
	}
	if approved.Status != domainwithdrawal.StatusApproved {
		t.Fatalf("status after approve = %q, want APPROVED", approved.Status)
	}

	sent, err := NewMarkWithdrawalSent(repo, nil).Execute(context.Background(), MarkWithdrawalSentInput{
		WithdrawalID:  "wd-1",
		ProviderRef:   "bank-ref-1",
		PayoutChannel: "PAYA",
		ReceiptURL:    "https://receipts.example/wd-1",
	})
	if err != nil {
		t.Fatalf("MarkWithdrawalSent() error = %v", err)
	}
	if sent.Status != domainwithdrawal.StatusSentToBank {
		t.Fatalf("status after sent = %q, want SENT_TO_BANK", sent.Status)
	}
	if sent.ProviderRef != "bank-ref-1" || sent.PayoutChannel != "PAYA" || sent.ReceiptURL == "" {
		t.Fatalf("sent receipt fields not persisted: %+v", sent)
	}

	settled, err := NewSettleWithdrawal(repo, nil).Execute(context.Background(), "wd-1")
	if err != nil {
		t.Fatalf("SettleWithdrawal() error = %v", err)
	}
	if settled.Status != domainwithdrawal.StatusSettled {
		t.Fatalf("status after settle = %q, want SETTLED", settled.Status)
	}
	if settled.ProcessedAt == nil {
		t.Fatal("settled withdrawal should set ProcessedAt")
	}
}

func TestCreateWithdrawalRejectsInactiveUser(t *testing.T) {
	uc := NewCreateWithdrawal(nil, nil, nil, nil, nil)
	_, err := uc.Execute(context.Background(), CreateWithdrawalInput{
		UserID:          "user-1",
		Amount:          1000,
		DestinationIBAN: "IR123",
		UserActive:      false,
	})
	if err == nil {
		t.Fatal("expected inactive user error")
	}
}
