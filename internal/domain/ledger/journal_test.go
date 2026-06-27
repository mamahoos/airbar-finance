package ledger

import (
	"testing"
)

func TestValidateLinesBalanced(t *testing.T) {
	lines := []EntryLine{
		{AccountCode: ShipmentEscrowAccount("sh-1"), Debit: 1000, Credit: 0},
		{AccountCode: UserWalletAccount("user-1"), Debit: 0, Credit: 1000},
	}

	if err := ValidateLines(lines); err != nil {
		t.Fatalf("ValidateLines() error = %v", err)
	}
}

func TestValidateLinesUnbalanced(t *testing.T) {
	lines := []EntryLine{
		{AccountCode: ShipmentEscrowAccount("sh-1"), Debit: 1000, Credit: 0},
		{AccountCode: UserWalletAccount("user-1"), Debit: 0, Credit: 500},
	}

	if err := ValidateLines(lines); err != ErrUnbalancedJournal {
		t.Fatalf("ValidateLines() error = %v, want ErrUnbalancedJournal", err)
	}
}

func TestValidateLinesInvalidSide(t *testing.T) {
	lines := []EntryLine{
		{AccountCode: AccountIRPSPClearing, Debit: 100, Credit: 100},
	}

	if err := ValidateLines(lines); err != ErrInvalidEntry {
		t.Fatalf("ValidateLines() error = %v, want ErrInvalidEntry", err)
	}
}

func TestUserWalletAccountFormat(t *testing.T) {
	got := UserWalletAccount("u42").String()
	want := "USER:u42:IRT:WALLET_LIABILITY"
	if got != want {
		t.Fatalf("UserWalletAccount() = %q, want %q", got, want)
	}
}

func TestShipmentEscrowAccountFormat(t *testing.T) {
	got := ShipmentEscrowAccount("sh-99").String()
	want := "SHIPMENT:sh-99:IRT:ESCROW"
	if got != want {
		t.Fatalf("ShipmentEscrowAccount() = %q, want %q", got, want)
	}
}
