package provider

import (
	"context"

	domainprovider "github.com/mamahoos/airbar-finance/internal/domain/provider"
)

// ListProviderEvents returns PSP audit events for admin operations.
type ListProviderEvents struct {
	events domainprovider.Repository
}

// NewListProviderEvents creates the provider event listing use case.
func NewListProviderEvents(events domainprovider.Repository) *ListProviderEvents {
	return &ListProviderEvents{events: events}
}

// Execute lists provider events with bounded pagination.
func (uc *ListProviderEvents) Execute(ctx context.Context, filter domainprovider.ListFilter) ([]domainprovider.Event, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return uc.events.List(ctx, filter)
}
