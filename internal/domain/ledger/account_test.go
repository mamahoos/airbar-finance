package ledger

import "testing"

func TestParseUserIDFromWalletAccount(t *testing.T) {
	userID, ok := ParseUserIDFromWalletAccount(UserWalletAccount("u42"))
	if !ok || userID != "u42" {
		t.Fatalf("ParseUserIDFromWalletAccount() = (%q, %v), want (u42, true)", userID, ok)
	}

	_, ok = ParseUserIDFromWalletAccount(AccountIRPSPClearing)
	if ok {
		t.Fatal("expected false for system account")
	}
}
