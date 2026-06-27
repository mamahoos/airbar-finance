package withdrawal

import "testing"

func TestHashDestinationNormalizesIBAN(t *testing.T) {
	a := HashDestination(" ir12 3456 ")
	b := HashDestination("IR123456")
	if a != b {
		t.Fatalf("hash mismatch: %q vs %q", a, b)
	}
	if len(a) != 64 {
		t.Fatalf("hash length = %d, want 64", len(a))
	}
}
