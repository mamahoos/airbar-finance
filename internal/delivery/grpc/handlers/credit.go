package handlers

import (
	"context"
	"time"

	domaincredit "github.com/mamahoos/airbar-finance/internal/domain/credit"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	credituc "github.com/mamahoos/airbar-finance/internal/usecase/credit"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreditHandler implements CreditService (UC-25).
type CreditHandler struct {
	financev1.UnimplementedCreditServiceServer
	grantCredit  *credituc.GrantCredit
	reverseGrant *credituc.ReverseCreditGrant
	getBalance   *credituc.GetBalance
	listGrants   *credituc.ListGrants
}

// NewCreditHandler creates a CreditService gRPC handler.
func NewCreditHandler(
	grantCredit *credituc.GrantCredit,
	reverseGrant *credituc.ReverseCreditGrant,
	getBalance *credituc.GetBalance,
	listGrants *credituc.ListGrants,
) *CreditHandler {
	return &CreditHandler{
		grantCredit:  grantCredit,
		reverseGrant: reverseGrant,
		getBalance:   getBalance,
		listGrants:   listGrants,
	}
}

func (h *CreditHandler) GrantCredit(ctx context.Context, req *financev1.GrantCreditRequest) (*financev1.CreditGrantResponse, error) {
	amount, err := credituc.ParseAmount(req.GetAmount())
	if err != nil {
		return nil, mapCreditError(domaincredit.ErrInvalidInput)
	}
	idempotencyKey := req.GetContext().GetIdempotencyKey()
	if idempotencyKey == "" {
		return nil, mapCreditError(domaincredit.ErrInvalidInput)
	}

	var expiresAt *time.Time
	if req.GetExpiresAt() != nil {
		t := req.GetExpiresAt().AsTime()
		expiresAt = &t
	}

	grant, err := h.grantCredit.Execute(ctx, credituc.GrantCreditInput{
		UserID:         req.GetUserId(),
		Amount:         amount,
		Reason:         req.GetReason(),
		CampaignRef:    req.GetCampaignRef(),
		ExpiresAt:      expiresAt,
		GrantedBy:      req.GetGrantedBy(),
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, mapCreditError(err)
	}
	return toCreditGrantResponse(grant), nil
}

func (h *CreditHandler) ReverseCreditGrant(ctx context.Context, req *financev1.ReverseCreditGrantRequest) (*financev1.CreditGrantResponse, error) {
	grant, err := h.reverseGrant.Execute(ctx, credituc.ReverseCreditInput{
		GrantID:       req.GetGrantId(),
		ReverseReason: req.GetReverseReason(),
		ReversedBy:    req.GetReversedBy(),
	})
	if err != nil {
		return nil, mapCreditError(err)
	}
	return toCreditGrantResponse(grant), nil
}

func (h *CreditHandler) GetCreditBalance(ctx context.Context, req *financev1.GetCreditBalanceRequest) (*financev1.CreditBalanceResponse, error) {
	balance, err := h.getBalance.Execute(ctx, req.GetUserId())
	if err != nil {
		return nil, mapCreditError(err)
	}
	return &financev1.CreditBalanceResponse{
		UserId:      req.GetUserId(),
		Currency:    domaincredit.CurrencyIRT,
		Balance:     credituc.FormatAmount(balance),
		AccountCode: domainledger.UserPromoCreditAccount(req.GetUserId()).String(),
	}, nil
}

func (h *CreditHandler) ListCreditGrants(ctx context.Context, req *financev1.ListCreditGrantsRequest) (*financev1.CreditGrantsResponse, error) {
	result, err := h.listGrants.Execute(ctx, req.GetUserId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapCreditError(err)
	}

	resp := &financev1.CreditGrantsResponse{
		Balance: credituc.FormatAmount(result.Balance),
		Items:   make([]*financev1.CreditGrantResponse, len(result.Grants)),
	}
	for i := range result.Grants {
		resp.Items[i] = toCreditGrantResponse(&result.Grants[i])
	}
	return resp, nil
}

func toCreditGrantResponse(grant *domaincredit.Grant) *financev1.CreditGrantResponse {
	resp := &financev1.CreditGrantResponse{
		Id:          grant.ID,
		UserId:      grant.UserID,
		Amount:      credituc.FormatAmount(grant.AmountRials),
		Reason:      grant.Reason,
		CampaignRef: grant.CampaignRef,
		Status:      string(grant.Status),
		GrantedBy:   grant.GrantedBy,
		CreatedAt:   timestamppb.New(grant.CreatedAt),
		ReverseReason: grant.ReverseReason,
		ReversedBy:    grant.ReversedBy,
	}
	if grant.ExpiresAt != nil {
		resp.ExpiresAt = timestamppb.New(*grant.ExpiresAt)
	}
	if grant.ReversedAt != nil {
		resp.ReversedAt = timestamppb.New(*grant.ReversedAt)
	}
	return resp
}
