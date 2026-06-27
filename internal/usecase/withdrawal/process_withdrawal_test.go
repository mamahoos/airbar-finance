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
	uc := NewProcessWithdrawal(repo)

	result, err := uc.Execute(context.Background(), ProcessWithdrawalInput{WithdrawalID: "wd-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != domainwithdrawal.StatusCompleted {
		t.Fatalf("status = %q, want COMPLETED", result.Status)
	}
}

func TestCreateWithdrawalRejectsInactiveUser(t *testing.T) {
	uc := NewCreateWithdrawal(nil, nil, nil, nil)
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
