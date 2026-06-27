package handlers

import (
	"context"
	"errors"

	domainrecon "github.com/mamahoos/airbar-finance/internal/domain/reconciliation"
	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	reconuc "github.com/mamahoos/airbar-finance/internal/usecase/reconciliation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ReconciliationHandler implements ReconciliationService (UC-21..23).
type ReconciliationHandler struct {
	financev1.UnimplementedReconciliationServiceServer
	runReconciliation *reconuc.RunReconciliation
	listRuns          *reconuc.ListReconciliationRuns
	getRun            *reconuc.GetReconciliationRun
}

// NewReconciliationHandler creates a ReconciliationService gRPC handler.
func NewReconciliationHandler(
	runReconciliation *reconuc.RunReconciliation,
	listRuns *reconuc.ListReconciliationRuns,
	getRun *reconuc.GetReconciliationRun,
) *ReconciliationHandler {
	return &ReconciliationHandler{
		runReconciliation: runReconciliation,
		listRuns:          listRuns,
		getRun:            getRun,
	}
}

func (h *ReconciliationHandler) RunReconciliation(ctx context.Context, _ *financev1.RunReconciliationRequest) (*financev1.ReconciliationRunResponse, error) {
	run, err := h.runReconciliation.Execute(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return toReconciliationRunResponse(run)
}

func (h *ReconciliationHandler) ListReconciliationRuns(ctx context.Context, _ *financev1.ListReconciliationRunsRequest) (*financev1.ReconciliationRunsResponse, error) {
	runs, err := h.listRuns.Execute(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	resp := &financev1.ReconciliationRunsResponse{
		Items: make([]*financev1.ReconciliationRunResponse, len(runs)),
	}
	for i, run := range runs {
		item, err := toReconciliationRunResponse(&run)
		if err != nil {
			return nil, status.Error(codes.Internal, "internal error")
		}
		resp.Items[i] = item
	}
	return resp, nil
}

func (h *ReconciliationHandler) GetReconciliationRun(ctx context.Context, req *financev1.GetReconciliationRunRequest) (*financev1.ReconciliationRunResponse, error) {
	run, err := h.getRun.Execute(ctx, req.GetRunId())
	if err != nil {
		return nil, mapReconciliationError(err)
	}
	return toReconciliationRunResponse(run)
}

func toReconciliationRunResponse(run *domainrecon.Run) (*financev1.ReconciliationRunResponse, error) {
	findings, err := structpb.NewStruct(run.Findings)
	if err != nil {
		return nil, err
	}

	resp := &financev1.ReconciliationRunResponse{
		Id:        run.ID,
		Status:    string(run.Status),
		Findings:  findings,
		StartedAt: timestamppb.New(run.StartedAt),
	}
	if run.CompletedAt != nil {
		resp.CompletedAt = timestamppb.New(*run.CompletedAt)
	}
	return resp, nil
}

func mapReconciliationError(err error) error {
	if errors.Is(err, domainrecon.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, "internal error")
}
