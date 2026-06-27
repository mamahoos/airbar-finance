package payment

import "testing"

func TestParseAmount(t *testing.T) {
	amount, err := ParseAmount("1500000")
	if err != nil {
		t.Fatalf("ParseAmount() error = %v", err)
	}
	if amount != 1500000 {
		t.Fatalf("amount = %d, want 1500000", amount)
	}

	if _, err := ParseAmount("0"); err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestCallbackURL(t *testing.T) {
	got := CallbackURL("http://localhost:8080")
	want := "http://localhost:8080/api/v1/zibal/callback"
	if got != want {
		t.Fatalf("CallbackURL() = %q, want %q", got, want)
	}
}
