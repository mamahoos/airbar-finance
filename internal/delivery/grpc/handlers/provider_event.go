package handlers

import (
	"context"

	domainprovider "github.com/mamahoos/airbar-finance/internal/domain/provider"
	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	provideruc "github.com/mamahoos/airbar-finance/internal/usecase/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProviderEventHandler implements ProviderEventService.
type ProviderEventHandler struct {
	financev1.UnimplementedProviderEventServiceServer
	listEvents *provideruc.ListProviderEvents
}

// NewProviderEventHandler creates a ProviderEventService gRPC handler.
func NewProviderEventHandler(listEvents *provideruc.ListProviderEvents) *ProviderEventHandler {
	return &ProviderEventHandler{listEvents: listEvents}
}

func (h *ProviderEventHandler) ListProviderEvents(ctx context.Context, req *financev1.ListProviderEventsRequest) (*financev1.ProviderEventsResponse, error) {
	events, total, err := h.listEvents.Execute(ctx, domainprovider.ListFilter{
		Provider:       req.GetProvider(),
		EventType:      domainprovider.EventType(req.GetEventType()),
		PaymentOrderID: req.GetPaymentOrderId(),
		Page:           int(req.GetPage()),
		Limit:          int(req.GetLimit()),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	resp := &financev1.ProviderEventsResponse{
		Items: make([]*financev1.ProviderEventResponse, len(events)),
		Total: total,
	}
	for i, event := range events {
		resp.Items[i] = &financev1.ProviderEventResponse{
			Id:             event.ID,
			Provider:       event.Provider,
			EventType:      string(event.EventType),
			PaymentOrderId: event.PaymentOrderID,
			PayloadHash:    event.PayloadHash,
			IdempotencyKey: event.IdempotencyKey,
			Processed:      event.Processed,
			CreatedAt:      timestamppb.New(event.CreatedAt),
		}
	}
	return resp, nil
}
