package zibal

import (
	"context"
	"testing"
)

func TestMockServerRequestVerify(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	ctx := context.Background()
	result, err := mock.Client().Request(ctx, RequestInput{
		Amount:      1500000,
		CallbackURL: "http://localhost/callback",
		Description: "test",
		OrderID:     "order-1",
	})
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}
	if result.TrackID == "" {
		t.Fatal("expected track id")
	}

	verify, err := mock.Client().Verify(ctx, result.TrackID)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if verify.Amount != 1500000 {
		t.Fatalf("amount = %d, want 1500000", verify.Amount)
	}
}
