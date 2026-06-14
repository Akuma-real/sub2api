package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCalculateOpenAIDualLoserBilling(t *testing.T) {
	tests := []struct {
		name      string
		input     OpenAIDualAttemptBillingInput
		wantBasis string
		wantCost  float64
	}{
		{
			name: "terminal usage uses max actual and provider floor",
			input: OpenAIDualAttemptBillingInput{
				BillingBasis:      OpenAIDualBillingBasisTerminalUsage,
				ActualAttemptCost: 0.004,
				ProviderCostFloor: 0.006,
			},
			wantBasis: OpenAIDualBillingBasisTerminalUsage,
			wantCost:  0.006,
		},
		{
			name: "partial usage includes estimated input and minimum fee",
			input: OpenAIDualAttemptBillingInput{
				BillingBasis:        OpenAIDualBillingBasisPartialUsage,
				PartialObservedCost: 0.001,
				EstimatedInputCost:  0.002,
				ProviderCostFloor:   0.001,
				MinAttemptFee:       0.005,
			},
			wantBasis: OpenAIDualBillingBasisPartialUsage,
			wantCost:  0.005,
		},
		{
			name: "dispatched no usage uses max estimate floor and minimum fee",
			input: OpenAIDualAttemptBillingInput{
				BillingBasis:       OpenAIDualBillingBasisNoUsage,
				EstimatedInputCost: 0.003,
				ProviderCostFloor:  0.004,
				MinAttemptFee:      0.002,
				UpstreamDispatched: true,
			},
			wantBasis: OpenAIDualBillingBasisNoUsage,
			wantCost:  0.004,
		},
		{
			name: "not dispatched is free",
			input: OpenAIDualAttemptBillingInput{
				BillingBasis:       OpenAIDualBillingBasisNotDispatched,
				EstimatedInputCost: 0.003,
				ProviderCostFloor:  0.004,
				MinAttemptFee:      0.002,
			},
			wantBasis: OpenAIDualBillingBasisNotDispatched,
			wantCost:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateOpenAIDualLoserBilling(tt.input)
			require.Equal(t, tt.wantBasis, got.BillingBasis)
			require.InDelta(t, tt.wantCost, got.BilledCost, 1e-12)
			require.InDelta(t, tt.wantCost, got.ProtectedCost, 1e-12)
		})
	}
}

func TestBuildOpenAICostBreakdownSnapshotIncludesAccelerationAndVIP(t *testing.T) {
	tier := OpenAIFastTierPriority
	vipMultiplier := 0.8
	apiKey := &APIKey{
		AccelerationSettings: APIKeyAccelerationSettings{
			FastMode:                   AccelerationFastModeForcePriority,
			DualProtectionEnabled:      true,
			DualFirstResponseTimeoutMS: 9000,
		},
	}
	result := &OpenAIForwardResult{
		ServiceTier: &tier,
		DualProtection: &OpenAIDualProtectionResult{
			Enabled:      true,
			AttemptCount: 2,
			ExtraCost:    0.003,
			Attempts: []OpenAIDualAttempt{
				{
					AttemptID:  OpenAIDualAttemptRolePrimary,
					Role:       OpenAIDualAttemptRolePrimary,
					Outcome:    OpenAIDualAttemptOutcomeWinner,
					Status:     OpenAIDualAttemptStatusCompleted,
					BilledCost: 0.01,
				},
				{
					AttemptID:    OpenAIDualAttemptRoleSecondary,
					Role:         OpenAIDualAttemptRoleSecondary,
					Outcome:      OpenAIDualAttemptOutcomeLoser,
					Status:       OpenAIDualAttemptStatusCanceled,
					BillingBasis: openAIDualStringPtr(OpenAIDualBillingBasisNoUsage),
					BilledCost:   0.003,
				},
			},
		},
	}
	cost := &CostBreakdown{
		TotalCost:                 0.013,
		ActualCost:                0.013,
		VIPDiscountableActualCost: 0.01,
		VIPProtectedActualCost:    0.003,
		BillingMode:               string(BillingModeToken),
	}

	snapshot := BuildOpenAICostBreakdownSnapshot(result, cost, apiKey, &Account{ID: 7, Type: AccountTypeAPIKey}, false)
	usageLog := &UsageLog{
		ActualCost:            0.011,
		DualProtectionEnabled: true,
		DualAttemptCount:      2,
		DualExtraCost:         0.003,
		VIPDiscountMultiplier: &vipMultiplier,
		VIPSavingsUSD:         0.002,
	}
	snapshot = EnrichOpenAICostBreakdownSnapshotFromUsageLog(snapshot, usageLog)

	fast, ok := snapshot["fast"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, AccelerationFastModeForcePriority, fast["mode"])
	require.Equal(t, OpenAIFastTierPriority, fast["service_tier"])
	dual, ok := snapshot["dual"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, dual["enabled"])
	require.Equal(t, 2, dual["attempt_count"])
	require.InDelta(t, 0.003, dual["extra_cost"], 1e-12)
	require.InDelta(t, 0.003, dual["secondary_cost"], 1e-12)
	vip, ok := snapshot["vip"].(map[string]any)
	require.True(t, ok)
	require.InDelta(t, 0.8, vip["discount_multiplier"], 1e-12)
	require.InDelta(t, 0.002, vip["savings_usd"], 1e-12)
	final, ok := snapshot["final"].(map[string]any)
	require.True(t, ok)
	require.InDelta(t, 0.011, final["actual_cost"], 1e-12)
}

func TestBuildOpenAICostBreakdownSnapshotDoesNotMarkConfiguredDualAsActive(t *testing.T) {
	apiKey := &APIKey{
		AccelerationSettings: APIKeyAccelerationSettings{
			DualProtectionEnabled:      true,
			DualFirstResponseTimeoutMS: 9000,
		},
	}

	snapshot := BuildOpenAICostBreakdownSnapshot(&OpenAIForwardResult{}, &CostBreakdown{}, apiKey, nil, false)

	dual, ok := snapshot["dual"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, dual["configured"])
	require.Equal(t, false, dual["enabled"])
	require.Equal(t, 0, dual["attempt_count"])
	require.InDelta(t, 0, dual["extra_cost"], 1e-12)
}

func TestOpenAIDualAttemptPlanResultForSuccessMatchesAttemptID(t *testing.T) {
	plan := NewOpenAIDualAttemptPlan(&APIKey{
		AccelerationSettings: APIKeyAccelerationSettings{DualProtectionEnabled: true},
	}, "req-dual-1", time.Now())

	attempts := []OpenAIDualAttempt{
		{
			AttemptID:  "primary-10-aaa",
			Role:       OpenAIDualAttemptRolePrimary,
			Status:     OpenAIDualAttemptStatusCompleted,
			BilledCost: 0.01,
		},
		{
			AttemptID:  "secondary-20-bbb",
			Role:       OpenAIDualAttemptRoleSecondary,
			Status:     OpenAIDualAttemptStatusCompleted,
			BilledCost: 0.02,
		},
	}

	result := plan.ResultForSuccess(attempts, "secondary-20-bbb")

	require.NotNil(t, result)
	require.Equal(t, 2, result.AttemptCount)
	require.Equal(t, OpenAIDualAttemptOutcomeLoser, result.Attempts[0].Outcome)
	require.Equal(t, OpenAIDualAttemptOutcomeWinner, result.Attempts[1].Outcome)
	require.InDelta(t, 0.01, result.ExtraCost, 1e-12)
}

func TestOpenAIDualUnsupportedResultRecordsSkippedAttempt(t *testing.T) {
	result := NewOpenAIDualUnsupportedResult(&APIKey{
		AccelerationSettings: APIKeyAccelerationSettings{DualProtectionEnabled: true},
	}, "req-ws", "/v1/realtime", "GET", OpenAIDualBillingBasisUnsupportedWS)

	require.NotNil(t, result)
	require.True(t, result.Enabled)
	require.Len(t, result.Attempts, 1)
	attempt := result.Attempts[0]
	require.Equal(t, OpenAIDualAttemptRoleSecondary, attempt.Role)
	require.Equal(t, OpenAIDualAttemptOutcomeSkipped, attempt.Outcome)
	require.Equal(t, OpenAIDualAttemptStatusCreated, attempt.Status)
	require.Equal(t, OpenAIDualBillingBasisNotDispatched, *attempt.BillingBasis)
	require.Equal(t, OpenAIDualBillingBasisUnsupportedWS, *attempt.CancelReason)
	require.InDelta(t, 0, result.ExtraCost, 1e-12)
}

func TestBuildOpenAIDualAttemptRecordCanceledLoserIsProtected(t *testing.T) {
	cancelReason := "context canceled"
	now := time.Now()
	attempt := BuildOpenAIDualAttemptRecord(OpenAIDualAttemptRecordInput{
		RequestID:            "req-dual-cancel",
		AttemptID:            "secondary-2-cancel",
		APIKeyID:             11,
		UserID:               22,
		Endpoint:             "/v1/responses",
		Method:               "post",
		Role:                 OpenAIDualAttemptRoleSecondary,
		Dispatched:           true,
		Winner:               false,
		HasPartialUsage:      true,
		EstimatedInputCost:   0.003,
		PartialObservedCost:  0.002,
		ProviderCostFloor:    0.004,
		MinAttemptFee:        0.006,
		UpstreamDispatchedAt: &now,
		CancelReason:         &cancelReason,
		Err:                  contextCanceledForOpenAIDualTest{},
		Now:                  now,
	})

	require.Equal(t, "secondary-2-cancel", attempt.AttemptID)
	require.Equal(t, OpenAIDualAttemptStatusCanceled, attempt.Status)
	require.Equal(t, OpenAIDualAttemptOutcomePending, attempt.Outcome)
	require.Equal(t, OpenAIDualBillingBasisPartialUsage, *attempt.BillingBasis)
	require.InDelta(t, 0.006, attempt.BilledCost, 1e-12)
	require.InDelta(t, 0.003, attempt.EstimatedCost, 1e-12)
}

func TestApplyOpenAIDualExtraCostToUsageCostProtectsLoserFromVIP(t *testing.T) {
	cost := &CostBreakdown{
		ActualCost: 0.01,
		TotalCost:  0.01,
	}
	ApplyOpenAIDualExtraCostToUsageCost(cost, &OpenAIDualProtectionResult{ExtraCost: 0.004})

	require.InDelta(t, 0.014, cost.ActualCost, 1e-12)
	require.InDelta(t, 0.014, cost.TotalCost, 1e-12)
	require.InDelta(t, 0.01, cost.VIPDiscountableActualCost, 1e-12)
	require.InDelta(t, 0.004, cost.VIPProtectedActualCost, 1e-12)
}

func TestNormalizeAPIKeyAccelerationSettingsClampsDualTimeout(t *testing.T) {
	low := NormalizeAPIKeyAccelerationSettings(&APIKeyAccelerationSettings{DualFirstResponseTimeoutMS: 1})
	require.Equal(t, AccelerationDualFirstTimeoutMinMS, low.DualFirstResponseTimeoutMS)

	high := NormalizeAPIKeyAccelerationSettings(&APIKeyAccelerationSettings{DualFirstResponseTimeoutMS: 120000})
	require.Equal(t, AccelerationDualFirstTimeoutMaxMS, high.DualFirstResponseTimeoutMS)
}

func openAIDualStringPtr(v string) *string {
	return &v
}

type contextCanceledForOpenAIDualTest struct{}

func (contextCanceledForOpenAIDualTest) Error() string {
	return "context canceled"
}
