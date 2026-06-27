package escrow

import "testing"

func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		status  Status
		pay     bool
		fund    bool
		deliv   bool
		freeze  bool
		release bool
		refund  bool
		partial bool
	}{
		{StatusCreated, true, true, false, false, false, false, false},
		{StatusFunded, false, false, true, false, false, true, true},
		{StatusDisputeWindow, false, false, false, true, true, true, true},
		{StatusLocked, false, false, false, true, false, false, false},
		{StatusFrozen, false, false, false, false, false, true, true},
		{StatusReleased, false, false, false, false, false, false, false},
		{StatusRefunded, false, false, false, false, false, false, false},
		{StatusPartiallyRefunded, false, false, false, false, true, true, true},
	}

	for _, tt := range tests {
		if got := tt.status.CanPayFromWallet(); got != tt.pay {
			t.Errorf("%s CanPayFromWallet() = %v, want %v", tt.status, got, tt.pay)
		}
		if got := tt.status.CanFund(); got != tt.fund {
			t.Errorf("%s CanFund() = %v, want %v", tt.status, got, tt.fund)
		}
		if got := tt.status.CanMarkDelivered(); got != tt.deliv {
			t.Errorf("%s CanMarkDelivered() = %v, want %v", tt.status, got, tt.deliv)
		}
		if got := tt.status.CanFreeze(); got != tt.freeze {
			t.Errorf("%s CanFreeze() = %v, want %v", tt.status, got, tt.freeze)
		}
		if got := tt.status.CanRelease(); got != tt.release {
			t.Errorf("%s CanRelease() = %v, want %v", tt.status, got, tt.release)
		}
		if got := tt.status.CanRefund(); got != tt.refund {
			t.Errorf("%s CanRefund() = %v, want %v", tt.status, got, tt.refund)
		}
		if got := tt.status.CanPartialRefund(); got != tt.partial {
			t.Errorf("%s CanPartialRefund() = %v, want %v", tt.status, got, tt.partial)
		}
	}
}

func TestCalcPlatformFee(t *testing.T) {
	if fee := CalcPlatformFee(10000, 10); fee != 1000 {
		t.Fatalf("fee = %d, want 1000", fee)
	}
	if fee := CalcPlatformFee(0, 10); fee != 0 {
		t.Fatalf("fee for zero amount = %d, want 0", fee)
	}
	if fee := CalcPlatformFee(5, 10); fee != 0 {
		t.Fatalf("fee for small amount = %d, want 0", fee)
	}
}
