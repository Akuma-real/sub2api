//go:build unit

package service

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const paymentFulfillmentAffiliateSchema = `
CREATE TABLE user_affiliates (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    aff_code TEXT NOT NULL UNIQUE,
    inviter_id INTEGER NULL REFERENCES users(id) ON DELETE SET NULL,
    aff_count INTEGER NOT NULL DEFAULT 0,
    aff_quota DECIMAL(20,8) NOT NULL DEFAULT 0,
    aff_history_quota DECIMAL(20,8) NOT NULL DEFAULT 0,
    aff_rebate_rate_percent DECIMAL(5,2),
    aff_code_custom BOOLEAN NOT NULL DEFAULT false,
    aff_frozen_quota DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_affiliate_ledger (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    amount DECIMAL(20,8) NOT NULL,
    source_user_id INTEGER NULL REFERENCES users(id) ON DELETE SET NULL,
    source_order_id INTEGER NULL REFERENCES payment_orders(id) ON DELETE SET NULL,
    frozen_until TIMESTAMP NULL,
    balance_after DECIMAL(20,8) NULL,
    aff_quota_after DECIMAL(20,8) NULL,
    aff_frozen_quota_after DECIMAL(20,8) NULL,
    aff_history_quota_after DECIMAL(20,8) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_payment_audit_logs_order_action_uniq
    ON payment_audit_logs(order_id, action);
`

type paymentFulfillmentTestProvider struct {
	key            string
	supportedTypes []payment.PaymentType
}

type paymentFulfillmentAffiliateRepo struct {
	client *dbent.Client
}

func newPaymentFulfillmentAffiliateRepo(client *dbent.Client) *paymentFulfillmentAffiliateRepo {
	return &paymentFulfillmentAffiliateRepo{client: client}
}

func (r *paymentFulfillmentAffiliateRepo) clientFromContext(ctx context.Context) *dbent.Client {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}
	return r.client
}

func (r *paymentFulfillmentAffiliateRepo) EnsureUserAffiliate(ctx context.Context, userID int64) (*AffiliateSummary, error) {
	rows, err := r.clientFromContext(ctx).QueryContext(ctx, `
		SELECT user_id, aff_code, inviter_id, aff_count, aff_quota, aff_frozen_quota, aff_history_quota,
		       aff_rebate_rate_percent, created_at, updated_at
		FROM user_affiliates
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, ErrAffiliateProfileNotFound
	}
	var summary AffiliateSummary
	var inviterID sql.NullInt64
	var customRate sql.NullFloat64
	requireNoScanErr := rows.Scan(
		&summary.UserID,
		&summary.AffCode,
		&inviterID,
		&summary.AffCount,
		&summary.AffQuota,
		&summary.AffFrozenQuota,
		&summary.AffHistoryQuota,
		&customRate,
		&summary.CreatedAt,
		&summary.UpdatedAt,
	)
	if requireNoScanErr != nil {
		return nil, requireNoScanErr
	}
	if inviterID.Valid {
		summary.InviterID = &inviterID.Int64
	}
	if customRate.Valid {
		summary.AffRebateRatePercent = &customRate.Float64
	}
	return &summary, rows.Err()
}

func (r *paymentFulfillmentAffiliateRepo) AccrueQuota(ctx context.Context, inviterID, inviteeUserID int64, amount float64, freezeHours int, sourceOrderID *int64) (bool, error) {
	client := r.clientFromContext(ctx)
	var res sql.Result
	var err error
	if freezeHours > 0 {
		res, err = client.ExecContext(ctx, `
			UPDATE user_affiliates
			SET aff_frozen_quota = aff_frozen_quota + ?, aff_history_quota = aff_history_quota + ?, updated_at = CURRENT_TIMESTAMP
			WHERE user_id = ?
		`, amount, amount, inviterID)
	} else {
		res, err = client.ExecContext(ctx, `
			UPDATE user_affiliates
			SET aff_quota = aff_quota + ?, aff_history_quota = aff_history_quota + ?, updated_at = CURRENT_TIMESTAMP
			WHERE user_id = ?
		`, amount, amount, inviterID)
	}
	if err != nil {
		return false, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return false, nil
	}

	sourceOrderArg := any(nil)
	if sourceOrderID != nil {
		sourceOrderArg = *sourceOrderID
	}
	if freezeHours > 0 {
		_, err = client.ExecContext(ctx, `
			INSERT INTO user_affiliate_ledger (user_id, action, amount, source_user_id, source_order_id, frozen_until, created_at, updated_at)
			VALUES (?, 'accrue', ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, inviterID, amount, inviteeUserID, sourceOrderArg, time.Now().Add(time.Duration(freezeHours)*time.Hour))
	} else {
		_, err = client.ExecContext(ctx, `
			INSERT INTO user_affiliate_ledger (user_id, action, amount, source_user_id, source_order_id, created_at, updated_at)
			VALUES (?, 'accrue', ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, inviterID, amount, inviteeUserID, sourceOrderArg)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *paymentFulfillmentAffiliateRepo) GetAccruedRebateFromInvitee(ctx context.Context, inviterID, inviteeUserID int64) (float64, error) {
	rows, err := r.clientFromContext(ctx).QueryContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM user_affiliate_ledger
		WHERE user_id = ? AND source_user_id = ? AND action = 'accrue'
	`, inviterID, inviteeUserID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return 0, rows.Err()
	}
	var total float64
	if err := rows.Scan(&total); err != nil {
		return 0, err
	}
	return total, rows.Err()
}

func (r *paymentFulfillmentAffiliateRepo) GetAffiliateByCode(context.Context, string) (*AffiliateSummary, error) {
	panic("unexpected GetAffiliateByCode call")
}
func (r *paymentFulfillmentAffiliateRepo) BindInviter(context.Context, int64, int64) (bool, error) {
	panic("unexpected BindInviter call")
}
func (r *paymentFulfillmentAffiliateRepo) ThawFrozenQuota(context.Context, int64) (float64, error) {
	panic("unexpected ThawFrozenQuota call")
}
func (r *paymentFulfillmentAffiliateRepo) TransferQuotaToBalance(context.Context, int64) (float64, float64, error) {
	panic("unexpected TransferQuotaToBalance call")
}
func (r *paymentFulfillmentAffiliateRepo) ListInvitees(context.Context, int64, int) ([]AffiliateInvitee, error) {
	panic("unexpected ListInvitees call")
}
func (r *paymentFulfillmentAffiliateRepo) UpdateUserAffCode(context.Context, int64, string) error {
	panic("unexpected UpdateUserAffCode call")
}
func (r *paymentFulfillmentAffiliateRepo) ResetUserAffCode(context.Context, int64) (string, error) {
	panic("unexpected ResetUserAffCode call")
}
func (r *paymentFulfillmentAffiliateRepo) SetUserRebateRate(context.Context, int64, *float64) error {
	panic("unexpected SetUserRebateRate call")
}
func (r *paymentFulfillmentAffiliateRepo) BatchSetUserRebateRate(context.Context, []int64, *float64) error {
	panic("unexpected BatchSetUserRebateRate call")
}
func (r *paymentFulfillmentAffiliateRepo) ListUsersWithCustomSettings(context.Context, AffiliateAdminFilter) ([]AffiliateAdminEntry, int64, error) {
	panic("unexpected ListUsersWithCustomSettings call")
}
func (r *paymentFulfillmentAffiliateRepo) ListAffiliateInviteRecords(context.Context, AffiliateRecordFilter) ([]AffiliateInviteRecord, int64, error) {
	panic("unexpected ListAffiliateInviteRecords call")
}
func (r *paymentFulfillmentAffiliateRepo) ListAffiliateRebateRecords(context.Context, AffiliateRecordFilter) ([]AffiliateRebateRecord, int64, error) {
	panic("unexpected ListAffiliateRebateRecords call")
}
func (r *paymentFulfillmentAffiliateRepo) ListAffiliateTransferRecords(context.Context, AffiliateRecordFilter) ([]AffiliateTransferRecord, int64, error) {
	panic("unexpected ListAffiliateTransferRecords call")
}
func (r *paymentFulfillmentAffiliateRepo) GetAffiliateUserOverview(context.Context, int64) (*AffiliateUserOverview, error) {
	panic("unexpected GetAffiliateUserOverview call")
}

func (p paymentFulfillmentTestProvider) Name() string        { return p.key }
func (p paymentFulfillmentTestProvider) ProviderKey() string { return p.key }
func (p paymentFulfillmentTestProvider) SupportedTypes() []payment.PaymentType {
	return p.supportedTypes
}
func (p paymentFulfillmentTestProvider) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	panic("unexpected call")
}
func (p paymentFulfillmentTestProvider) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	panic("unexpected call")
}
func (p paymentFulfillmentTestProvider) VerifyNotification(ctx context.Context, rawBody string, headers map[string]string) (*payment.PaymentNotification, error) {
	panic("unexpected call")
}
func (p paymentFulfillmentTestProvider) Refund(ctx context.Context, req payment.RefundRequest) (*payment.RefundResponse, error) {
	panic("unexpected call")
}

// ---------------------------------------------------------------------------
// resolveRedeemAction — pure idempotency decision logic
// ---------------------------------------------------------------------------

func TestResolveRedeemAction_CodeNotFound(t *testing.T) {
	t.Parallel()
	action := resolveRedeemAction(nil, nil)
	assert.Equal(t, redeemActionCreate, action, "nil code with nil error should create")
}

func TestResolveRedeemAction_LookupError(t *testing.T) {
	t.Parallel()
	action := resolveRedeemAction(nil, errors.New("db connection lost"))
	assert.Equal(t, redeemActionCreate, action, "lookup error should fall back to create")
}

func TestResolveRedeemAction_LookupErrorWithNonNilCode(t *testing.T) {
	t.Parallel()
	// Edge case: both code and error are non-nil (shouldn't happen in practice,
	// but the function should still treat error as authoritative)
	code := &RedeemCode{Status: StatusUnused}
	action := resolveRedeemAction(code, errors.New("partial error"))
	assert.Equal(t, redeemActionCreate, action, "non-nil error should always result in create regardless of code")
}

func TestResolveRedeemAction_CodeExistsAndUsed(t *testing.T) {
	t.Parallel()
	code := &RedeemCode{
		Code:   "test-code-123",
		Status: StatusUsed,
		Type:   RedeemTypeBalance,
		Value:  10.0,
	}
	action := resolveRedeemAction(code, nil)
	assert.Equal(t, redeemActionSkipCompleted, action, "used code should skip to completed")
}

func TestResolveRedeemAction_CodeExistsAndUnused(t *testing.T) {
	t.Parallel()
	code := &RedeemCode{
		Code:   "test-code-456",
		Status: StatusUnused,
		Type:   RedeemTypeBalance,
		Value:  25.0,
	}
	action := resolveRedeemAction(code, nil)
	assert.Equal(t, redeemActionRedeem, action, "unused code should skip creation and proceed to redeem")
}

func TestResolveRedeemAction_CodeExistsWithExpiredStatus(t *testing.T) {
	t.Parallel()
	// A code with a non-standard status (neither "unused" nor "used")
	// should NOT be treated as used, so it falls through to redeemActionRedeem.
	code := &RedeemCode{
		Code:   "expired-code",
		Status: StatusExpired,
	}
	action := resolveRedeemAction(code, nil)
	assert.Equal(t, redeemActionRedeem, action, "expired-status code is not IsUsed(), should redeem")
}

// ---------------------------------------------------------------------------
// Table-driven comprehensive test
// ---------------------------------------------------------------------------

func TestResolveRedeemAction_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     *RedeemCode
		err      error
		expected redeemAction
	}{
		{
			name:     "nil code, nil error — first run",
			code:     nil,
			err:      nil,
			expected: redeemActionCreate,
		},
		{
			name:     "nil code, lookup error — treat as not found",
			code:     nil,
			err:      ErrRedeemCodeNotFound,
			expected: redeemActionCreate,
		},
		{
			name:     "nil code, generic DB error — treat as not found",
			code:     nil,
			err:      errors.New("connection refused"),
			expected: redeemActionCreate,
		},
		{
			name:     "code exists, used — previous run completed redeem",
			code:     &RedeemCode{Status: StatusUsed},
			err:      nil,
			expected: redeemActionSkipCompleted,
		},
		{
			name:     "code exists, unused — previous run created code but crashed before redeem",
			code:     &RedeemCode{Status: StatusUnused},
			err:      nil,
			expected: redeemActionRedeem,
		},
		{
			name:     "code exists but error also set — error takes precedence",
			code:     &RedeemCode{Status: StatusUsed},
			err:      errors.New("unexpected"),
			expected: redeemActionCreate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := resolveRedeemAction(tt.code, tt.err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// ---------------------------------------------------------------------------
// redeemAction enum value sanity
// ---------------------------------------------------------------------------

func TestRedeemAction_DistinctValues(t *testing.T) {
	t.Parallel()
	// Ensure the three actions have distinct values (iota correctness)
	assert.NotEqual(t, redeemActionCreate, redeemActionRedeem)
	assert.NotEqual(t, redeemActionCreate, redeemActionSkipCompleted)
	assert.NotEqual(t, redeemActionRedeem, redeemActionSkipCompleted)
}

// ---------------------------------------------------------------------------
// RedeemCode.IsUsed / CanUse interaction with resolveRedeemAction
// ---------------------------------------------------------------------------

func TestResolveRedeemAction_IsUsedCanUseConsistency(t *testing.T) {
	t.Parallel()

	usedCode := &RedeemCode{Status: StatusUsed}
	unusedCode := &RedeemCode{Status: StatusUnused}

	// Verify our decision function is consistent with the domain model methods
	assert.True(t, usedCode.IsUsed())
	assert.False(t, usedCode.CanUse())
	assert.Equal(t, redeemActionSkipCompleted, resolveRedeemAction(usedCode, nil))

	assert.False(t, unusedCode.IsUsed())
	assert.True(t, unusedCode.CanUse())
	assert.Equal(t, redeemActionRedeem, resolveRedeemAction(unusedCode, nil))
}

func TestPaymentOrderCanAccrueAffiliateRebateIncludesSubscriptionOrders(t *testing.T) {
	t.Parallel()

	assert.True(t, paymentOrderCanAccrueAffiliateRebate(&dbent.PaymentOrder{
		OrderType: payment.OrderTypeBalance,
		Amount:    10,
	}))
	assert.True(t, paymentOrderCanAccrueAffiliateRebate(&dbent.PaymentOrder{
		OrderType: payment.OrderTypeSubscription,
		Amount:    99,
	}))
	assert.False(t, paymentOrderCanAccrueAffiliateRebate(&dbent.PaymentOrder{
		OrderType: payment.OrderTypeSubscription,
		Amount:    0,
	}))
	assert.False(t, paymentOrderCanAccrueAffiliateRebate(&dbent.PaymentOrder{
		OrderType: "other",
		Amount:    99,
	}))
}

func TestExecuteSubscriptionFulfillmentAccruesAffiliateRebate(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	inviter, err := client.User.Create().
		SetEmail("sub-aff-inviter@example.com").
		SetPasswordHash("hash").
		SetUsername("sub-aff-inviter").
		Save(ctx)
	require.NoError(t, err)
	invitee, err := client.User.Create().
		SetEmail("sub-aff-invitee@example.com").
		SetPasswordHash("hash").
		SetUsername("sub-aff-invitee").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ExecContext(ctx, paymentFulfillmentAffiliateSchema)
	require.NoError(t, err)
	_, err = client.ExecContext(ctx, `
		INSERT INTO user_affiliates (user_id, aff_code, inviter_id, aff_count, aff_quota, aff_frozen_quota, aff_history_quota, created_at, updated_at)
		VALUES (?, ?, NULL, 1, 0, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		       (?, ?, ?, 0, 0, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, inviter.ID, "SUBINVITER01", invitee.ID, "SUBINVITEE01", inviter.ID)
	require.NoError(t, err)

	groupID := int64(101)
	days := 30
	order, err := client.PaymentOrder.Create().
		SetUserID(invitee.ID).
		SetUserEmail(invitee.Email).
		SetUserName(invitee.Username).
		SetAmount(99).
		SetPayAmount(99).
		SetFeeRate(0).
		SetRechargeCode("SUB-AFF-ORDER").
		SetOutTradeNo("sub2_sub_aff_order").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("sub-aff-trade").
		SetOrderType(payment.OrderTypeSubscription).
		SetSubscriptionGroupID(groupID).
		SetSubscriptionDays(days).
		SetStatus(OrderStatusPaid).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	settingSvc := NewSettingService(&paymentConfigSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:           "true",
		SettingKeyAffiliateRebateRate:        "20",
		SettingKeyAffiliateRebateFreezeHours: "0",
	}}, nil)
	affiliateSvc := NewAffiliateService(newPaymentFulfillmentAffiliateRepo(client), settingSvc, nil, nil)
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: groupID, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	subSvc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)
	svc := &PaymentService{
		entClient:        client,
		groupRepo:        groupRepo,
		subscriptionSvc:  subSvc,
		affiliateService: affiliateSvc,
	}

	require.NoError(t, svc.ExecuteSubscriptionFulfillment(ctx, order.ID))

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, reloaded.Status)
	require.Equal(t, 1, subRepo.createCalls)

	rows, err := client.QueryContext(ctx, `
		SELECT aff_quota, aff_history_quota
		FROM user_affiliates
		WHERE user_id = ?
	`, inviter.ID)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	require.True(t, rows.Next())
	var quota, history float64
	require.NoError(t, rows.Scan(&quota, &history))
	require.InDelta(t, 19.8, quota, 1e-9)
	require.InDelta(t, 19.8, history, 1e-9)

	logs, err := svc.GetOrderAuditLogs(ctx, order.ID)
	require.NoError(t, err)
	var sawRebate, sawSubscription bool
	for _, log := range logs {
		if log.Action == "AFFILIATE_REBATE_APPLIED" {
			sawRebate = true
			require.Contains(t, log.Detail, `"rebateAmount":19.8`)
		}
		if log.Action == "SUBSCRIPTION_SUCCESS" {
			sawSubscription = true
		}
	}
	require.True(t, sawRebate)
	require.True(t, sawSubscription)
}

func TestExpectedNotificationProviderKeyPrefersOrderInstanceProvider(t *testing.T) {
	t.Parallel()

	registry := payment.NewRegistry()
	registry.Register(paymentFulfillmentTestProvider{
		key:            payment.TypeAlipay,
		supportedTypes: []payment.PaymentType{payment.TypeAlipay},
	})

	assert.Equal(t,
		payment.TypeEasyPay,
		expectedNotificationProviderKey(registry, payment.TypeAlipay, "", payment.TypeEasyPay),
	)
}

func TestExpectedNotificationProviderKeyUsesRegistryMappingForLegacyOrders(t *testing.T) {
	t.Parallel()

	registry := payment.NewRegistry()
	registry.Register(paymentFulfillmentTestProvider{
		key:            payment.TypeEasyPay,
		supportedTypes: []payment.PaymentType{payment.TypeAlipay},
	})

	assert.Equal(t,
		payment.TypeEasyPay,
		expectedNotificationProviderKey(registry, payment.TypeAlipay, "", ""),
	)
}

func TestExpectedNotificationProviderKeyFallsBackToPaymentType(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		payment.TypeWxpay,
		expectedNotificationProviderKey(nil, payment.TypeWxpay, "", ""),
	)
}

func TestExpectedNotificationProviderKeyPrefersOrderSnapshotProviderKey(t *testing.T) {
	t.Parallel()

	registry := payment.NewRegistry()
	registry.Register(paymentFulfillmentTestProvider{
		key:            payment.TypeAlipay,
		supportedTypes: []payment.PaymentType{payment.TypeAlipay},
	})

	assert.Equal(t,
		payment.TypeEasyPay,
		expectedNotificationProviderKey(registry, payment.TypeAlipay, payment.TypeEasyPay, ""),
	)
}

func TestExpectedNotificationProviderKeyForOrderUsesSnapshotProviderKey(t *testing.T) {
	t.Parallel()

	registry := payment.NewRegistry()
	registry.Register(paymentFulfillmentTestProvider{
		key:            payment.TypeAlipay,
		supportedTypes: []payment.PaymentType{payment.TypeAlipay},
	})

	order := &dbent.PaymentOrder{
		PaymentType: payment.TypeAlipay,
		ProviderSnapshot: map[string]any{
			"schema_version": 1,
			"provider_key":   payment.TypeEasyPay,
		},
	}

	assert.Equal(t,
		payment.TypeEasyPay,
		expectedNotificationProviderKeyForOrder(registry, order, ""),
	)
}

func TestValidateProviderNotificationMetadataRejectsWxpaySnapshotMismatch(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		PaymentType: payment.TypeWxpay,
		ProviderSnapshot: map[string]any{
			"schema_version":  1,
			"merchant_app_id": "wx-app-expected",
			"merchant_id":     "mch-expected",
			"currency":        "CNY",
		},
	}

	err := validateProviderNotificationMetadata(order, payment.TypeWxpay, map[string]string{
		"appid":       "wx-app-other",
		"mchid":       "mch-expected",
		"currency":    "CNY",
		"trade_state": "SUCCESS",
	})
	assert.ErrorContains(t, err, "wxpay appid mismatch")
}

func TestValidateProviderNotificationMetadataAllowsLegacyOrdersWithoutSnapshotFields(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		PaymentType: payment.TypeWxpay,
		ProviderSnapshot: map[string]any{
			"schema_version":       1,
			"provider_instance_id": "9",
			"provider_key":         payment.TypeWxpay,
		},
	}

	err := validateProviderNotificationMetadata(order, payment.TypeWxpay, map[string]string{
		"appid":       "wx-app-runtime",
		"mchid":       "mch-runtime",
		"currency":    "CNY",
		"trade_state": "SUCCESS",
	})
	assert.NoError(t, err)
}

func TestParseLegacyPaymentOrderID(t *testing.T) {
	t.Parallel()

	oid, ok := parseLegacyPaymentOrderID("sub2_42", &dbent.NotFoundError{})
	assert.True(t, ok)
	assert.EqualValues(t, 42, oid)

	_, ok = parseLegacyPaymentOrderID("42", &dbent.NotFoundError{})
	assert.False(t, ok)

	_, ok = parseLegacyPaymentOrderID("sub2_42", errors.New("db down"))
	assert.False(t, ok)
}

func TestIsValidProviderAmount(t *testing.T) {
	t.Parallel()

	assert.True(t, isValidProviderAmount(0.01))
	assert.False(t, isValidProviderAmount(0))
	assert.False(t, isValidProviderAmount(-1))
	assert.False(t, isValidProviderAmount(math.NaN()))
	assert.False(t, isValidProviderAmount(math.Inf(1)))
}

func TestValidateProviderNotificationMetadataRejectsAlipaySnapshotMismatch(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		PaymentType: payment.TypeAlipay,
		ProviderSnapshot: map[string]any{
			"schema_version":  2,
			"merchant_app_id": "alipay-app-expected",
		},
	}

	err := validateProviderNotificationMetadata(order, payment.TypeAlipay, map[string]string{
		"app_id": "alipay-app-other",
	})
	assert.ErrorContains(t, err, "alipay app_id mismatch")
}

func TestValidateProviderNotificationMetadataRejectsEasyPaySnapshotMismatch(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		PaymentType: payment.TypeAlipay,
		ProviderSnapshot: map[string]any{
			"schema_version": 2,
			"merchant_id":    "pid-expected",
		},
	}

	err := validateProviderNotificationMetadata(order, payment.TypeEasyPay, map[string]string{
		"pid": "pid-other",
	})
	assert.ErrorContains(t, err, "easypay pid mismatch")
}

func TestValidateProviderNotificationMetadataRejectsAirwallexSnapshotMismatch(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		PaymentType: payment.TypeAirwallex,
		ProviderSnapshot: map[string]any{
			"schema_version": 2,
			"merchant_id":    "acct_expected",
			"currency":       "CNY",
		},
	}

	err := validateProviderNotificationMetadata(order, payment.TypeAirwallex, map[string]string{
		"account_id": "acct_other",
		"currency":   "CNY",
		"status":     "SUCCEEDED",
	})
	assert.ErrorContains(t, err, "airwallex account_id mismatch")

	err = validateProviderNotificationMetadata(order, payment.TypeAirwallex, map[string]string{
		"account_id": "acct_expected",
		"currency":   "USD",
		"status":     "SUCCEEDED",
	})
	assert.ErrorContains(t, err, "airwallex currency mismatch")
}

func TestValidateProviderNotificationMetadataRejectsStripeCurrencyMismatch(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		PaymentType: payment.TypeStripe,
		ProviderSnapshot: map[string]any{
			"schema_version": 2,
			"currency":       "HKD",
		},
	}

	err := validateProviderNotificationMetadata(order, payment.TypeStripe, map[string]string{
		"currency": "USD",
	})
	assert.ErrorContains(t, err, "stripe currency mismatch")
}

func TestPaymentAmountToleranceForThreeDecimalCurrency(t *testing.T) {
	t.Parallel()

	assert.Equal(t, amountToleranceCNY, paymentAmountToleranceForCurrency("CNY"))
	assert.Equal(t, amountToleranceCNY, paymentAmountToleranceForCurrency("JPY"))
	assert.InDelta(t, 0.0005, paymentAmountToleranceForCurrency("KWD"), 1e-12)
}
