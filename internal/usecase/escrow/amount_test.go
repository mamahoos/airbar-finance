package escrow

import "testing"

func TestParseAmount(t *testing.T) {
	amount, err := ParseAmount("10000")
	if err != nil {
		t.Fatalf("ParseAmount() error = %v", err)
	}
	if amount != 10000 {
		t.Fatalf("amount = %d, want 10000", amount)
	}

	if _, err := ParseAmount("0"); err == nil {
		t.Fatal("expected error for zero amount")
	}
	if _, err := ParseAmount("abc"); err == nil {
		t.Fatal("expected error for invalid amount")
	}
}
