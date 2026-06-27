package idempotency_test

import (
	"testing"

	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	grpcidempotency "github.com/mamahoos/airbar-finance/internal/delivery/grpc/idempotency"
)

func TestSnapshotRoundTripEscrowResponse(t *testing.T) {
	const method = "/airbar.finance.v1.EscrowService/CreateEscrow"

	original := &financev1.EscrowResponse{
		Id:         "esc-1",
		ShipmentId: "sh-1",
		Status:     "CREATED",
		Amount:     "10000",
	}

	snapshot, err := grpcidempotency.ResponseToSnapshot(method, original)
	if err != nil {
		t.Fatalf("ResponseToSnapshot() error = %v", err)
	}

	replayed := &financev1.EscrowResponse{}
	if err := grpcidempotency.SnapshotToResponse(method, snapshot, replayed); err != nil {
		t.Fatalf("SnapshotToResponse() error = %v", err)
	}

	if replayed.GetId() != original.GetId() || replayed.GetStatus() != original.GetStatus() {
		t.Fatalf("replayed = %#v, want %#v", replayed, original)
	}
}
