//go:build unit

package service

import (
	"context"
	"strconv"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestBuildPaymentOrderProviderSnapshot_ExcludesSensitiveConfig(t *testing.T) {
	t.Parallel()

	sel := &payment.InstanceSelection{
		InstanceID:     "12",
		ProviderKey:    payment.TypeWxpay,
		SupportedTypes: "wxpay,wxpay_direct",
		PaymentMode:    "popup",
		Config: map[string]string{
			"privateKey": "secret",
			"apiV3Key":   "secret-v3",
			"appId":      "wx-app-id",
		},
	}

	snapshot := buildPaymentOrderProviderSnapshot(sel, CreateOrderRequest{})
	require.Equal(t, map[string]any{
		"schema_version":       2,
		"provider_instance_id": "12",
		"provider_key":         payment.TypeWxpay,
		"payment_mode":         "popup",
		"merchant_app_id":      "wx-app-id",
		"currency":             "CNY",
	}, snapshot)
	require.NotContains(t, snapshot, "config")
	require.NotContains(t, snapshot, "privateKey")
	require.NotContains(t, snapshot, "apiV3Key")
	require.NotContains(t, snapshot, "supported_types")
	require.NotContains(t, snapshot, "instance_name")
	require.NotContains(t, snapshot, "merchant_id")
}

func TestCreateOrderInTx_WritesProviderSnapshot(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("snapshot@example.com").
		SetPasswordHash("hash").
		SetUsername("snapshot-user").
		Save(ctx)
	require.NoError(t, err)

	instance, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName("Primary Alipay").
		SetConfig(`{"secretKey":"do-not-copy"}`).
		SetSupportedTypes("alipay,alipay_direct").
		SetPaymentMode("redirect").
		SetEnabled(true).
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}
	order, err := svc.createOrderInTx(
		ctx,
		CreateOrderRequest{
			UserID:      user.ID,
			PaymentType: payment.TypeAlipay,
			OrderType:   payment.OrderTypeBalance,
			ClientIP:    "127.0.0.1",
			SrcHost:     "app.example.com",
		},
		&User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		},
		nil,
		nil,
		&PaymentConfig{
			MaxPendingOrders: 3,
			OrderTimeoutMin:  30,
		},
		88,
		88,
		0,
		88,
		&payment.InstanceSelection{
			InstanceID:     strconv.FormatInt(instance.ID, 10),
			ProviderKey:    payment.TypeAlipay,
			SupportedTypes: "alipay,alipay_direct",
			PaymentMode:    "redirect",
			Config: map[string]string{
				"secretKey": "do-not-copy",
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, strconv.FormatInt(instance.ID, 10), valueOrEmpty(order.ProviderInstanceID))
	require.Equal(t, payment.TypeAlipay, valueOrEmpty(order.ProviderKey))
	require.Equal(t, float64(2), order.ProviderSnapshot["schema_version"])
	require.Equal(t, strconv.FormatInt(instance.ID, 10), order.ProviderSnapshot["provider_instance_id"])
	require.Equal(t, payment.TypeAlipay, order.ProviderSnapshot["provider_key"])
	require.Equal(t, "redirect", order.ProviderSnapshot["payment_mode"])
	require.NotContains(t, order.ProviderSnapshot, "config")
	require.NotContains(t, order.ProviderSnapshot, "secretKey")
	require.NotContains(t, order.ProviderSnapshot, "supported_types")
	require.NotContains(t, order.ProviderSnapshot, "instance_name")
}

func TestBuildPaymentOrderProviderSnapshot_UsesWxpayJSAPIAppIDForOpenIDOrders(t *testing.T) {
	t.Parallel()

	snapshot := buildPaymentOrderProviderSnapshot(&payment.InstanceSelection{
		InstanceID:  "88",
		ProviderKey: payment.TypeWxpay,
		Config: map[string]string{
			"appId":   "wx-open-app",
			"mpAppId": "wx-mp-app",
			"mchId":   "mch-88",
		},
		PaymentMode: "jsapi",
	}, CreateOrderRequest{OpenID: "openid-123"})

	require.Equal(t, "wx-mp-app", snapshot["merchant_app_id"])
	require.Equal(t, "mch-88", snapshot["merchant_id"])
	require.Equal(t, "CNY", snapshot["currency"])
}

func TestBuildPaymentOrderProviderSnapshot_IncludesAlipayMerchantIdentity(t *testing.T) {
	t.Parallel()

	snapshot := buildPaymentOrderProviderSnapshot(&payment.InstanceSelection{
		InstanceID:  "21",
		ProviderKey: payment.TypeAlipay,
		Config: map[string]string{
			"appId":      "alipay-app-21",
			"privateKey": "secret",
		},
		PaymentMode: "redirect",
	}, CreateOrderRequest{})

	require.Equal(t, "alipay-app-21", snapshot["merchant_app_id"])
	require.NotContains(t, snapshot, "privateKey")
}

func TestBuildPaymentOrderProviderSnapshot_IncludesEasyPayMerchantIdentity(t *testing.T) {
	t.Parallel()

	snapshot := buildPaymentOrderProviderSnapshot(&payment.InstanceSelection{
		InstanceID:  "66",
		ProviderKey: payment.TypeEasyPay,
		Config: map[string]string{
			"pid":  "easypay-merchant-66",
			"pkey": "secret",
		},
		PaymentMode: "popup",
	}, CreateOrderRequest{PaymentType: payment.TypeAlipay})

	require.Equal(t, "easypay-merchant-66", snapshot["merchant_id"])
	require.NotContains(t, snapshot, "pkey")
}

func TestBuildPaymentOrderProviderSnapshot_IncludesMuyinPlatformAndChannel(t *testing.T) {
	t.Parallel()

	snapshot := buildPaymentOrderProviderSnapshot(&payment.InstanceSelection{
		InstanceID:  "88",
		ProviderKey: payment.TypeMuyin,
		Config: map[string]string{
			"token":         "secret-token",
			"platform":      "merchant-platform",
			"alipayChannel": "FACE_TO_FACE_PAYMENT",
		},
		PaymentMode: "redirect",
	}, CreateOrderRequest{PaymentType: payment.TypeAlipay})

	require.Equal(t, "merchant-platform", snapshot["platform"])
	require.Equal(t, "FACE_TO_FACE_PAYMENT", snapshot["payment_channel"])
	require.NotContains(t, snapshot, "token")

	wxpaySnapshot := buildPaymentOrderProviderSnapshot(&payment.InstanceSelection{
		InstanceID:  "89",
		ProviderKey: payment.TypeMuyin,
		Config:      map[string]string{},
	}, CreateOrderRequest{PaymentType: payment.TypeWxpay})
	require.NotContains(t, wxpaySnapshot, "platform")
	require.Equal(t, "WECHATPAY_H5", wxpaySnapshot["payment_channel"])
}

func TestValidateProviderSnapshotMetadataChecksMuyinPaymentID(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		OutTradeNo:     "sub2_order",
		PaymentTradeNo: "pay_123",
		ProviderSnapshot: map[string]any{
			"schema_version": 2,
			"provider_key":   payment.TypeMuyin,
		},
	}

	require.NoError(t, validateProviderSnapshotMetadata(order, payment.TypeMuyin, map[string]string{
		"payment_id": "pay_123",
		"status":     "SUCCESS",
	}))
	require.ErrorContains(t, validateProviderSnapshotMetadata(order, payment.TypeMuyin, map[string]string{
		"payment_id": "pay_other",
		"status":     "SUCCESS",
	}), "payment_id mismatch")
	require.ErrorContains(t, validateProviderSnapshotMetadata(order, payment.TypeMuyin, map[string]string{
		"payment_id": "pay_123",
		"status":     "FAILED",
	}), "status mismatch")
}

func TestValidateProviderSnapshotMetadataChecksMuyinTypeChannelAndPlatform(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		OutTradeNo:     "sub2_order",
		PaymentTradeNo: "pay_123",
		PaymentType:    payment.TypeWxpay,
		ProviderSnapshot: map[string]any{
			"schema_version":  2,
			"provider_key":    payment.TypeMuyin,
			"payment_channel": "WECHATPAY_H5",
			"platform":        "test-muyin-user",
		},
	}
	goodMetadata := map[string]string{
		"payment_id":      "pay_123",
		"payment_type":    "WECHATPAY",
		"payment_channel": "WECHATPAY_H5",
		"platform":        "test-muyin-user",
		"status":          "SUCCESS",
	}

	require.NoError(t, validateProviderSnapshotMetadata(order, payment.TypeMuyin, goodMetadata))

	badType := cloneStringMap(goodMetadata)
	badType["payment_type"] = "ALIPAY"
	require.ErrorContains(t, validateProviderSnapshotMetadata(order, payment.TypeMuyin, badType), "payment_type mismatch")

	badChannel := cloneStringMap(goodMetadata)
	badChannel["payment_channel"] = "WECHATPAY_JSAPI"
	require.ErrorContains(t, validateProviderSnapshotMetadata(order, payment.TypeMuyin, badChannel), "payment_channel mismatch")

	badPlatform := cloneStringMap(goodMetadata)
	badPlatform["platform"] = "other-platform"
	require.ErrorContains(t, validateProviderSnapshotMetadata(order, payment.TypeMuyin, badPlatform), "platform mismatch")
}

func TestBuildPaymentOrderProviderSnapshot_IncludesProviderCurrency(t *testing.T) {
	t.Parallel()

	stripeSnapshot := buildPaymentOrderProviderSnapshot(&payment.InstanceSelection{
		InstanceID:  "77",
		ProviderKey: payment.TypeStripe,
		Config: map[string]string{
			"currency": "hkd",
		},
	}, CreateOrderRequest{})
	require.Equal(t, "HKD", stripeSnapshot["currency"])

	airwallexSnapshot := buildPaymentOrderProviderSnapshot(&payment.InstanceSelection{
		InstanceID:  "78",
		ProviderKey: payment.TypeAirwallex,
		Config: map[string]string{
			"currency":  "usd",
			"accountId": "acct-78",
		},
	}, CreateOrderRequest{})
	require.Equal(t, "USD", airwallexSnapshot["currency"])
	require.Equal(t, "acct-78", airwallexSnapshot["merchant_id"])
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
