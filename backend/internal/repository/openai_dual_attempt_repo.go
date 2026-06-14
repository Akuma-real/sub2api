package repository

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/openaidualattempt"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type openAIDualAttemptRepository struct {
	client *dbent.Client
}

func NewOpenAIDualAttemptRepository(client *dbent.Client) service.OpenAIDualAttemptRepository {
	return &openAIDualAttemptRepository{client: client}
}

func (r *openAIDualAttemptRepository) Upsert(ctx context.Context, attempt *service.OpenAIDualAttempt) error {
	if r == nil || r.client == nil || attempt == nil {
		return nil
	}
	attempt.Normalize()
	builder := r.client.OpenAIDualAttempt.Create().
		SetRequestID(attempt.RequestID).
		SetAttemptID(attempt.AttemptID).
		SetAPIKeyID(attempt.APIKeyID).
		SetUserID(attempt.UserID).
		SetEndpoint(attempt.Endpoint).
		SetMethod(attempt.Method).
		SetRole(attempt.Role).
		SetOutcome(attempt.Outcome).
		SetStatus(attempt.Status).
		SetEstimatedCost(attempt.EstimatedCost).
		SetActualCost(attempt.ActualCost).
		SetBilledCost(attempt.BilledCost).
		SetMetadata(attempt.Metadata)
	if attempt.AccountID != nil {
		builder.SetAccountID(*attempt.AccountID)
	}
	if attempt.ServiceTier != nil && strings.TrimSpace(*attempt.ServiceTier) != "" {
		builder.SetServiceTier(strings.TrimSpace(*attempt.ServiceTier))
	}
	if attempt.BillingBasis != nil && strings.TrimSpace(*attempt.BillingBasis) != "" {
		builder.SetBillingBasis(strings.TrimSpace(*attempt.BillingBasis))
	}
	if attempt.UpstreamDispatchedAt != nil {
		builder.SetUpstreamDispatchedAt(*attempt.UpstreamDispatchedAt)
	}
	if attempt.CancelReason != nil && strings.TrimSpace(*attempt.CancelReason) != "" {
		builder.SetCancelReason(strings.TrimSpace(*attempt.CancelReason))
	}

	return builder.
		OnConflict(
			sql.ConflictColumns(
				openaidualattempt.FieldRequestID,
				openaidualattempt.FieldAPIKeyID,
				openaidualattempt.FieldAttemptID,
			),
		).
		UpdateNewValues().
		Exec(ctx)
}

func (r *openAIDualAttemptRepository) ListByRequest(ctx context.Context, requestID string, apiKeyID int64) ([]service.OpenAIDualAttempt, error) {
	if r == nil || r.client == nil {
		return nil, nil
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" || apiKeyID <= 0 {
		return []service.OpenAIDualAttempt{}, nil
	}
	rows, err := r.client.OpenAIDualAttempt.Query().
		Where(
			openaidualattempt.RequestIDEQ(requestID),
			openaidualattempt.APIKeyIDEQ(apiKeyID),
		).
		Order(dbent.Asc(openaidualattempt.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]service.OpenAIDualAttempt, 0, len(rows))
	for _, row := range rows {
		out = append(out, openAIDualAttemptEntityToService(row))
	}
	return out, nil
}

func openAIDualAttemptEntityToService(row *dbent.OpenAIDualAttempt) service.OpenAIDualAttempt {
	if row == nil {
		return service.OpenAIDualAttempt{}
	}
	return service.OpenAIDualAttempt{
		ID:                   row.ID,
		RequestID:            row.RequestID,
		AttemptID:            row.AttemptID,
		APIKeyID:             row.APIKeyID,
		UserID:               row.UserID,
		AccountID:            row.AccountID,
		Endpoint:             row.Endpoint,
		Method:               row.Method,
		Role:                 row.Role,
		Outcome:              row.Outcome,
		ServiceTier:          row.ServiceTier,
		Status:               row.Status,
		BillingBasis:         row.BillingBasis,
		EstimatedCost:        row.EstimatedCost,
		ActualCost:           row.ActualCost,
		BilledCost:           row.BilledCost,
		UpstreamDispatchedAt: row.UpstreamDispatchedAt,
		CancelReason:         row.CancelReason,
		Metadata:             row.Metadata,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}
