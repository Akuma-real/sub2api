package service

import (
	"context"
	"math"
	"strings"
	"time"
)

type OpenAIDualAttemptBillingInput struct {
	BillingBasis        string
	ActualAttemptCost   float64
	PartialObservedCost float64
	EstimatedInputCost  float64
	ProviderCostFloor   float64
	MinAttemptFee       float64
	UpstreamDispatched  bool
}

type OpenAIDualAttemptBillingResult struct {
	BillingBasis  string
	BilledCost    float64
	ProtectedCost float64
}

func CalculateOpenAIDualLoserBilling(input OpenAIDualAttemptBillingInput) OpenAIDualAttemptBillingResult {
	basis := input.BillingBasis
	if basis == "" {
		if !input.UpstreamDispatched {
			basis = OpenAIDualBillingBasisNotDispatched
		} else {
			basis = OpenAIDualBillingBasisNoUsage
		}
	}
	var billed float64
	switch basis {
	case OpenAIDualBillingBasisTerminalUsage:
		billed = maxFloat64(input.ActualAttemptCost, input.ProviderCostFloor)
	case OpenAIDualBillingBasisPartialUsage:
		billed = maxFloat64(input.PartialObservedCost+input.EstimatedInputCost, input.ProviderCostFloor, input.MinAttemptFee)
	case OpenAIDualBillingBasisNotDispatched:
		billed = 0
	default:
		basis = OpenAIDualBillingBasisNoUsage
		billed = maxFloat64(input.EstimatedInputCost, input.ProviderCostFloor, input.MinAttemptFee)
	}
	if math.IsNaN(billed) || math.IsInf(billed, 0) || billed < 0 {
		billed = 0
	}
	return OpenAIDualAttemptBillingResult{
		BillingBasis:  basis,
		BilledCost:    billed,
		ProtectedCost: billed,
	}
}

func BuildOpenAIDualAttemptRecord(input OpenAIDualAttemptRecordInput) OpenAIDualAttempt {
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	role := normalizeOpenAIDualAttemptRole(input.Role)
	outcome := OpenAIDualAttemptOutcomePending
	if input.Winner {
		outcome = OpenAIDualAttemptOutcomeWinner
	} else if !input.Dispatched && input.Err != nil {
		outcome = OpenAIDualAttemptOutcomeSkipped
	}
	status := OpenAIDualAttemptStatusCreated
	if input.Dispatched && input.Err != nil {
		status = OpenAIDualAttemptStatusFailed
		if input.CancelReason != nil && strings.Contains(strings.ToLower(strings.TrimSpace(*input.CancelReason)), "cancel") {
			status = OpenAIDualAttemptStatusCanceled
		}
	} else if input.Dispatched {
		status = OpenAIDualAttemptStatusCompleted
	}
	basis := OpenAIDualBillingBasisNotDispatched
	actualCost := sanitizeNonNegativeFloat(input.ActualAttemptCost)
	billedCost := 0.0
	if input.Dispatched {
		if input.HasTerminalUsage {
			basis = OpenAIDualBillingBasisTerminalUsage
		} else if input.HasPartialUsage {
			basis = OpenAIDualBillingBasisPartialUsage
		} else {
			basis = OpenAIDualBillingBasisNoUsage
		}
		if input.Winner {
			billedCost = actualCost
		} else {
			billing := CalculateOpenAIDualLoserBilling(OpenAIDualAttemptBillingInput{
				BillingBasis:        basis,
				ActualAttemptCost:   actualCost,
				PartialObservedCost: input.PartialObservedCost,
				EstimatedInputCost:  input.EstimatedInputCost,
				ProviderCostFloor:   input.ProviderCostFloor,
				MinAttemptFee:       input.MinAttemptFee,
				UpstreamDispatched:  true,
			})
			basis = billing.BillingBasis
			billedCost = billing.BilledCost
		}
	}
	attemptID := role
	if trimmed := strings.TrimSpace(input.AttemptID); trimmed != "" {
		attemptID = trimmed
	}
	attempt := OpenAIDualAttempt{
		RequestID:            strings.TrimSpace(input.RequestID),
		AttemptID:            attemptID,
		APIKeyID:             input.APIKeyID,
		UserID:               input.UserID,
		AccountID:            input.AccountID,
		Endpoint:             strings.TrimSpace(input.Endpoint),
		Method:               strings.TrimSpace(input.Method),
		Role:                 role,
		Outcome:              outcome,
		ServiceTier:          input.ServiceTier,
		Status:               status,
		BillingBasis:         &basis,
		EstimatedCost:        sanitizeNonNegativeFloat(input.EstimatedInputCost),
		ActualCost:           actualCost,
		BilledCost:           sanitizeNonNegativeFloat(billedCost),
		UpstreamDispatchedAt: input.UpstreamDispatchedAt,
		CancelReason:         input.CancelReason,
		Metadata:             input.Metadata,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if attempt.Metadata == nil {
		attempt.Metadata = map[string]any{}
	}
	if input.Err != nil {
		attempt.Metadata["error"] = strings.TrimSpace(input.Err.Error())
	}
	if attempt.CancelReason != nil {
		reason := trimOpenAIDualAttemptReason(*attempt.CancelReason)
		attempt.CancelReason = &reason
	}
	attempt.Normalize()
	return attempt
}

type OpenAIDualAttemptCostInput struct {
	Result          *OpenAIForwardResult
	APIKey          *APIKey
	BillingModels   []string
	Multiplier      float64
	ImageMultiplier float64
}

func (s *OpenAIGatewayService) CalculateOpenAIAttemptCostForDual(ctx context.Context, input OpenAIDualAttemptCostInput) (*CostBreakdown, error) {
	if s == nil || input.Result == nil || input.APIKey == nil {
		return nil, nil
	}
	actualInputTokens := input.Result.Usage.InputTokens - input.Result.Usage.CacheReadInputTokens
	if actualInputTokens < 0 {
		actualInputTokens = 0
	}
	tokens := UsageTokens{
		InputTokens:         actualInputTokens,
		OutputTokens:        input.Result.Usage.OutputTokens,
		CacheCreationTokens: input.Result.Usage.CacheCreationInputTokens,
		CacheReadTokens:     input.Result.Usage.CacheReadInputTokens,
		ImageOutputTokens:   input.Result.Usage.ImageOutputTokens,
	}
	serviceTier := ""
	if input.Result.ServiceTier != nil {
		serviceTier = strings.TrimSpace(*input.Result.ServiceTier)
	}
	return s.calculateOpenAIRecordUsageCost(ctx, input.Result, input.APIKey, input.BillingModels, input.Multiplier, input.ImageMultiplier, tokens, serviceTier)
}

func ApplyOpenAIDualExtraCostToUsageCost(cost *CostBreakdown, dual *OpenAIDualProtectionResult) {
	if cost == nil || dual == nil || dual.ExtraCost <= 0 {
		return
	}
	extra := sanitizeNonNegativeFloat(dual.ExtraCost)
	if cost.VIPDiscountableActualCost <= 0 && cost.VIPProtectedActualCost <= 0 {
		cost.VIPDiscountableActualCost = cost.ActualCost
	}
	cost.ActualCost += extra
	cost.TotalCost += extra
	cost.VIPProtectedActualCost += extra
}

type OpenAIDualAttemptRecordInput struct {
	RequestID            string
	AttemptID            string
	APIKeyID             int64
	UserID               int64
	AccountID            *int64
	Endpoint             string
	Method               string
	Role                 string
	ServiceTier          *string
	Dispatched           bool
	Winner               bool
	HasTerminalUsage     bool
	HasPartialUsage      bool
	EstimatedInputCost   float64
	PartialObservedCost  float64
	ActualAttemptCost    float64
	ProviderCostFloor    float64
	MinAttemptFee        float64
	UpstreamDispatchedAt *time.Time
	CancelReason         *string
	Metadata             map[string]any
	Err                  error
	Now                  time.Time
}

func sanitizeNonNegativeFloat(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
		return 0
	}
	return v
}

func maxFloat64(values ...float64) float64 {
	var max float64
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			continue
		}
		if value > max {
			max = value
		}
	}
	return max
}

func trimOpenAIDualAttemptReason(v string) string {
	v = strings.TrimSpace(v)
	if len(v) <= 128 {
		return v
	}
	return v[:128]
}

func EffectiveOpenAIDualProtectionEnabled(apiKey *APIKey) bool {
	if apiKey == nil {
		return false
	}
	return apiKey.AccelerationSettings.Normalize().DualProtectionEnabled
}

func OpenAIDualFirstResponseTimeout(apiKey *APIKey) int {
	settings := DefaultAPIKeyAccelerationSettings()
	if apiKey != nil {
		settings = apiKey.AccelerationSettings.Normalize()
	}
	if settings.DualFirstResponseTimeoutMS <= 0 {
		return AccelerationDualFirstTimeoutMS
	}
	return settings.DualFirstResponseTimeoutMS
}

func BuildOpenAICostBreakdownSnapshot(result *OpenAIForwardResult, cost *CostBreakdown, apiKey *APIKey, account *Account, isSubscriptionBill bool) map[string]any {
	snapshot := map[string]any{
		"fast": map[string]any{
			"mode":         apiKeyAccelerationFastMode(apiKey),
			"service_tier": optionalStringValue(resultServiceTier(result)),
		},
		"dual": map[string]any{
			"configured":         EffectiveOpenAIDualProtectionEnabled(apiKey),
			"enabled":            false,
			"first_timeout_ms":   OpenAIDualFirstResponseTimeout(apiKey),
			"attempt_count":      0,
			"primary_cost":       0,
			"secondary_cost":     0,
			"extra_cost":         0,
			"unsupported_reason": "",
			"billing_disclaimer": "all dispatched upstream attempts may be billed, including losers not returned to the client",
		},
		"vip": map[string]any{
			"discount_multiplier": nil,
			"savings_usd":         0,
		},
		"final": map[string]any{
			"billing_type": "balance",
			"actual_cost":  0,
		},
	}
	if isSubscriptionBill {
		snapshot["final"].(map[string]any)["billing_type"] = "subscription"
	}
	if cost != nil {
		snapshot["base"] = map[string]any{
			"input_cost":              cost.InputCost,
			"output_cost":             cost.OutputCost,
			"cache_creation_cost":     cost.CacheCreationCost,
			"cache_read_cost":         cost.CacheReadCost,
			"image_output_cost":       cost.ImageOutputCost,
			"total_cost":              cost.TotalCost,
			"rate_multiplier_applied": cost.ActualCost,
			"billing_mode":            cost.BillingMode,
		}
		snapshot["final"].(map[string]any)["actual_cost"] = cost.ActualCost
		if cost.VIPProtectedActualCost > 0 || cost.VIPDiscountableActualCost > 0 {
			snapshot["vip"].(map[string]any)["discountable_cost"] = cost.VIPDiscountableActualCost
			snapshot["vip"].(map[string]any)["protected_cost"] = cost.VIPProtectedActualCost
		}
	}
	if account != nil {
		snapshot["account"] = map[string]any{
			"id":              account.ID,
			"type":            account.Type,
			"rate_multiplier": account.BillingRateMultiplier(),
		}
	}
	if result != nil && result.DualProtection != nil {
		dual := snapshot["dual"].(map[string]any)
		dual["enabled"] = result.DualProtection.Enabled
		dual["attempt_count"] = result.DualProtection.AttemptCount
		dual["extra_cost"] = result.DualProtection.ExtraCost
		attempts := make([]map[string]any, 0, len(result.DualProtection.Attempts))
		for _, attempt := range result.DualProtection.Attempts {
			entry := map[string]any{
				"attempt_id":     attempt.AttemptID,
				"role":           attempt.Role,
				"outcome":        attempt.Outcome,
				"status":         attempt.Status,
				"billing_basis":  optionalStringValue(attempt.BillingBasis),
				"estimated_cost": attempt.EstimatedCost,
				"actual_cost":    attempt.ActualCost,
				"billed_cost":    attempt.BilledCost,
			}
			if attempt.CancelReason != nil {
				entry["cancel_reason"] = strings.TrimSpace(*attempt.CancelReason)
				if strings.TrimSpace(*attempt.CancelReason) == OpenAIDualBillingBasisUnsupportedWS {
					dual["unsupported_reason"] = OpenAIDualBillingBasisUnsupportedWS
				}
			}
			attempts = append(attempts, entry)
			switch attempt.Role {
			case OpenAIDualAttemptRoleSecondary:
				dual["secondary_cost"] = attempt.BilledCost
			default:
				dual["primary_cost"] = attempt.BilledCost
			}
		}
		dual["attempts"] = attempts
	}
	return snapshot
}

func EnrichOpenAICostBreakdownSnapshotFromUsageLog(snapshot map[string]any, usageLog *UsageLog) map[string]any {
	if snapshot == nil {
		snapshot = map[string]any{}
	}
	if usageLog == nil {
		return snapshot
	}
	dual, _ := snapshot["dual"].(map[string]any)
	if dual == nil {
		dual = map[string]any{}
		snapshot["dual"] = dual
	}
	dual["enabled"] = usageLog.DualProtectionEnabled
	dual["attempt_count"] = usageLog.DualAttemptCount
	dual["extra_cost"] = usageLog.DualExtraCost

	vip, _ := snapshot["vip"].(map[string]any)
	if vip == nil {
		vip = map[string]any{}
		snapshot["vip"] = vip
	}
	if usageLog.VIPLevelID != nil {
		vip["level_id"] = *usageLog.VIPLevelID
	}
	if usageLog.VIPDiscountMultiplier != nil {
		vip["discount_multiplier"] = *usageLog.VIPDiscountMultiplier
	}
	if usageLog.VIPPreDiscountCost != nil {
		vip["pre_discount_cost"] = *usageLog.VIPPreDiscountCost
	}
	vip["savings_usd"] = usageLog.VIPSavingsUSD

	final, _ := snapshot["final"].(map[string]any)
	if final == nil {
		final = map[string]any{}
		snapshot["final"] = final
	}
	final["actual_cost"] = usageLog.ActualCost
	return snapshot
}

func apiKeyAccelerationFastMode(apiKey *APIKey) string {
	if apiKey == nil {
		return AccelerationFastModeOff
	}
	return apiKey.AccelerationSettings.Normalize().FastMode
}

func resultServiceTier(result *OpenAIForwardResult) *string {
	if result == nil {
		return nil
	}
	return result.ServiceTier
}

func optionalStringValue(v *string) any {
	if v == nil {
		return nil
	}
	return strings.TrimSpace(*v)
}

func OpenAIUsageHasAnyUsage(usage OpenAIUsage) bool {
	return usage.InputTokens > 0 ||
		usage.OutputTokens > 0 ||
		usage.CacheCreationInputTokens > 0 ||
		usage.CacheReadInputTokens > 0 ||
		usage.ImageOutputTokens > 0
}

func EstimateOpenAIInputCost(cost *CostBreakdown) float64 {
	if cost == nil {
		return 0
	}
	return sanitizeNonNegativeFloat(cost.InputCost + cost.CacheCreationCost + cost.CacheReadCost)
}
