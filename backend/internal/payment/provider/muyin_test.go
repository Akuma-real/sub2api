//go:build unit

package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

const testMuyinPlatform = "test-muyin-user"

func TestNewMuyinValidatesConfigAndDefaults(t *testing.T) {
	t.Parallel()

	_, err := NewMuyin("1", map[string]string{})
	require.ErrorContains(t, err, "token")

	_, err = NewMuyin("1", map[string]string{
		"token":     "tok",
		"notifyUrl": "https://merchant.example.com/api/v1/payment/webhook/muyin",
		"returnUrl": "https://merchant.example.com/payment/result",
	})
	require.ErrorContains(t, err, "platform")

	prov, err := NewMuyin("1", map[string]string{
		"token":     "tok",
		"notifyUrl": "https://merchant.example.com/api/v1/payment/webhook/muyin",
		"returnUrl": "https://merchant.example.com/payment/result",
		"platform":  testMuyinPlatform,
	})
	require.NoError(t, err)
	require.Equal(t, payment.TypeMuyin, prov.ProviderKey())
	require.Equal(t, []payment.PaymentType{payment.TypeAlipay, payment.TypeWxpay}, prov.SupportedTypes())
	require.Equal(t, muyinDefaultAPIBase, prov.config["apiBase"])
	require.Equal(t, testMuyinPlatform, prov.config["platform"])
	require.Equal(t, muyinAlipayDefaultChannel, prov.config["alipayChannel"])
	require.Equal(t, muyinWxpayDefaultChannel, prov.config["wxpayChannel"])
}

func TestMuyinCreatePaymentSendsDocumentedPayload(t *testing.T) {
	t.Parallel()

	var gotPayload muyinPaymentRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/apis/platform.payment.muyin.site/v1alpha1/initiatePayment", r.URL.Path)
		require.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
		require.Contains(t, r.Header.Get("Content-Type"), "application/json")
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &gotPayload))
		_, _ = w.Write([]byte(`{
			"orderId":"sub2_order",
			"paymentId":"pay_123",
			"paymentType":"ALIPAY",
			"paymentUrl":"https://qr.alipay.com/xxx",
			"qrCode":{"qrCodeContent":"qr-content"},
			"status":"UNPAID",
			"amount":12.34,
			"message":"success"
		}`))
	}))
	defer server.Close()

	prov := mustTestMuyinProvider(t, server)
	resp, err := prov.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:     "sub2_order",
		Amount:      "12.34",
		PaymentType: payment.TypeAlipay,
		Subject:     "Sub2API 12.34 CNY",
	})
	require.NoError(t, err)
	require.Equal(t, "pay_123", resp.TradeNo)
	require.Equal(t, "https://qr.alipay.com/xxx", resp.PayURL)
	require.Equal(t, "qr-content", resp.QRCode)
	require.Equal(t, "Sub2API 12.34 CNY", gotPayload.Subject)
	require.Equal(t, 12.34, gotPayload.Amount)
	require.Equal(t, muyinAlipayType, gotPayload.PaymentType)
	require.Equal(t, muyinAlipayDefaultChannel, gotPayload.PaymentChannel)
	require.Equal(t, "sub2_order", gotPayload.OrderID)
	require.Equal(t, "https://merchant.example.com/api/v1/payment/webhook/muyin", gotPayload.NotifyURL)
	require.Equal(t, "https://merchant.example.com/payment/result", gotPayload.ReturnURL)
	require.Equal(t, testMuyinPlatform, gotPayload.Platform)
}

func TestDecodeMuyinPaymentResultPrefersWrappedData(t *testing.T) {
	t.Parallel()

	result, err := decodeMuyinPaymentResult([]byte(`{
		"message":"ok",
		"data":{
			"orderId":"sub2_order",
			"paymentId":"pay_123",
			"paymentType":"ALIPAY",
			"status":"SUCCESS",
			"amount":12.34
		}
	}`))
	require.NoError(t, err)
	require.Equal(t, "sub2_order", result.OrderID)
	require.Equal(t, "pay_123", result.PaymentID)
	require.Equal(t, "ALIPAY", result.PaymentType)
	require.Equal(t, "SUCCESS", result.Status)
	require.Equal(t, 12.34, result.Amount)
	require.Equal(t, "ok", result.Message)
}

func TestMuyinSupportedPaymentChannelsCallsDocumentedEndpoint(t *testing.T) {
	t.Parallel()

	var gotPayload muyinPaymentRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/apis/platform.payment.muyin.site/v1alpha1/supportedPaymentChannels", r.URL.Path)
		require.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
		_, _ = w.Write([]byte(`[{"label":"支付宝-当面付","value":"FACE_TO_FACE_PAYMENT"}]`))
	}))
	defer server.Close()

	prov := mustTestMuyinProvider(t, server)
	channels, err := prov.SupportedPaymentChannels(context.Background(), payment.TypeAlipay)
	require.NoError(t, err)
	require.Equal(t, muyinAlipayType, gotPayload.PaymentType)
	require.Equal(t, testMuyinPlatform, gotPayload.Platform)
	require.Equal(t, []MuyinPaymentChannelOption{
		{Label: "支付宝-当面付", Value: "FACE_TO_FACE_PAYMENT"},
	}, channels)
}

func TestMuyinCreateWxpayH5AddsClientIPAdditionalData(t *testing.T) {
	t.Parallel()

	var gotPayload muyinPaymentRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &gotPayload))
		_, _ = w.Write([]byte(`{
			"orderId":"sub2_wx_order",
			"paymentId":"pay_wx_123",
			"paymentType":"WECHATPAY",
			"paymentUrl":"https://wx.example.com/pay",
			"status":"UNPAID",
			"amount":20
		}`))
	}))
	defer server.Close()

	prov := mustTestMuyinProvider(t, server)
	_, err := prov.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:     "sub2_wx_order",
		Amount:      "20",
		PaymentType: payment.TypeWxpay,
		Subject:     "Sub2API 20 CNY",
		ClientIP:    "203.0.113.10",
	})
	require.NoError(t, err)
	require.Equal(t, muyinWechatPayType, gotPayload.PaymentType)
	require.Equal(t, muyinWxpayDefaultChannel, gotPayload.PaymentChannel)
	require.Equal(t, "203.0.113.10", gotPayload.ClientIP)
	require.NotNil(t, gotPayload.AdditionalData)
	require.Equal(t, "203.0.113.10", gotPayload.AdditionalData.KeyValues["payerClientIp"])
	require.Equal(t, "Wap", gotPayload.AdditionalData.KeyValues["h5Type"])
}

func TestMuyinQueryOrderMapsStatuses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apis/platform.payment.muyin.site/v1alpha1/queryPaymentByOrderId":
			var payload muyinPaymentRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			require.Equal(t, "sub2_order", payload.OrderID)
			require.Equal(t, muyinAlipayType, payload.PaymentType)
			_, _ = w.Write([]byte(`{
				"orderId":"sub2_order",
				"paymentId":"pay_123",
				"paymentType":"ALIPAY",
				"status":"SUCCESS",
				"tradeNo":"trade_123",
				"amount":12.34,
				"paymentTime":"2026-05-28T12:00:00Z"
			}`))
		case "/apis/platform.payment.muyin.site/v1alpha1/listPaymentInfo":
			require.Equal(t, "sub2_order", r.URL.Query().Get("orderId"))
			require.Equal(t, "pay_123", r.URL.Query().Get("paymentId"))
			_, _ = w.Write([]byte(`{"items":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	prov := mustTestMuyinProvider(t, server)
	resp, err := prov.QueryOrder(context.Background(), "sub2_order")
	require.NoError(t, err)
	require.Equal(t, payment.ProviderStatusPaid, resp.Status)
	require.Equal(t, "pay_123", resp.TradeNo)
	require.Equal(t, 12.34, resp.Amount)
	require.Equal(t, "2026-05-28T12:00:00Z", resp.PaidAt)
	require.Equal(t, "ALIPAY", resp.Metadata["payment_type"])
}

func TestMuyinQueryOrderUsesPaymentTypeFromContext(t *testing.T) {
	t.Parallel()

	var gotPayloads []muyinPaymentRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apis/platform.payment.muyin.site/v1alpha1/queryPaymentByOrderId":
			var payload muyinPaymentRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			gotPayloads = append(gotPayloads, payload)
			require.Equal(t, "sub2_wx_order", payload.OrderID)
			require.Equal(t, muyinWechatPayType, payload.PaymentType)
			_, _ = w.Write([]byte(`{
				"orderId":"sub2_wx_order",
				"paymentId":"pay_wx_123",
				"paymentType":"WECHATPAY",
				"status":"SUCCESS",
				"amount":20
			}`))
		case "/apis/platform.payment.muyin.site/v1alpha1/listPaymentInfo":
			_, _ = w.Write([]byte(`{"items":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	prov := mustTestMuyinProvider(t, server)
	queryCtx := payment.WithQueryPaymentType(context.Background(), payment.TypeWxpay)
	resp, err := prov.QueryOrder(queryCtx, "sub2_wx_order")
	require.NoError(t, err)
	require.Equal(t, payment.ProviderStatusPaid, resp.Status)
	require.Equal(t, "pay_wx_123", resp.TradeNo)
	require.Len(t, gotPayloads, 1)
	require.Equal(t, muyinWechatPayType, gotPayloads[0].PaymentType)
}

func TestMuyinQueryOrderMergesPaymentInfoMetadata(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apis/platform.payment.muyin.site/v1alpha1/queryPaymentByOrderId":
			_, _ = w.Write([]byte(`{
				"orderId":"sub2_order",
				"paymentId":"pay_123",
				"paymentType":"ALIPAY",
				"status":"SUCCESS",
				"amount":12.34
			}`))
		case "/apis/platform.payment.muyin.site/v1alpha1/listPaymentInfo":
			require.Equal(t, "sub2_order", r.URL.Query().Get("orderId"))
			require.Equal(t, "pay_123", r.URL.Query().Get("paymentId"))
			require.Equal(t, testMuyinPlatform, r.URL.Query().Get("platform"))
			_, _ = w.Write([]byte(`{"data":{"items":[{
				"orderId":"sub2_order",
				"paymentId":"pay_123",
				"paymentType":"ALIPAY",
				"paymentChannel":"FACE_TO_FACE_PAYMENT",
				"paymentStatus":"SUCCESS",
				"platform":"test-muyin-user"
			}]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	prov := mustTestMuyinProvider(t, server)
	resp, err := prov.QueryOrder(context.Background(), "sub2_order")
	require.NoError(t, err)
	require.Equal(t, "FACE_TO_FACE_PAYMENT", resp.Metadata["payment_channel"])
	require.Equal(t, testMuyinPlatform, resp.Metadata["platform"])
	require.Equal(t, "SUCCESS", resp.Metadata["status"])
}

func TestMuyinQueryOrderUsesPaymentAmountFromPaymentInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apis/platform.payment.muyin.site/v1alpha1/queryPaymentByOrderId":
			_, _ = w.Write([]byte(`{
				"orderId":"sub2_order",
				"paymentId":"pay_123",
				"paymentType":"ALIPAY",
				"status":"SUCCESS"
			}`))
		case "/apis/platform.payment.muyin.site/v1alpha1/listPaymentInfo":
			_, _ = w.Write([]byte(`{"items":[{
				"orderId":"sub2_order",
				"paymentId":"pay_123",
				"paymentType":"ALIPAY",
				"paymentStatus":"SUCCESS",
				"paymentAmount":"12.34"
			}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	prov := mustTestMuyinProvider(t, server)
	resp, err := prov.QueryOrder(context.Background(), "sub2_order")
	require.NoError(t, err)
	require.Equal(t, payment.ProviderStatusPaid, resp.Status)
	require.Equal(t, 12.34, resp.Amount)
}

func TestMuyinVerifyNotificationActivelyQueriesPayment(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apis/platform.payment.muyin.site/v1alpha1/queryPaymentByPaymentId":
			var payload muyinPaymentRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			require.Equal(t, "pay_123", payload.PaymentID)
			require.Equal(t, muyinAlipayType, payload.PaymentType)
			_, _ = w.Write([]byte(`{
				"orderId":"sub2_order",
				"paymentId":"pay_123",
				"paymentType":"ALIPAY",
				"status":"SUCCESS",
				"amount":12.34
			}`))
		case "/apis/platform.payment.muyin.site/v1alpha1/listPaymentInfo":
			require.Equal(t, "pay_123", r.URL.Query().Get("paymentId"))
			_, _ = w.Write([]byte(`{"items":[{
				"orderId":"sub2_order",
				"paymentId":"pay_123",
				"paymentType":"ALIPAY",
				"paymentChannel":"FACE_TO_FACE_PAYMENT",
				"paymentStatus":"SUCCESS",
				"platform":"test-muyin-user"
			}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	prov := mustTestMuyinProvider(t, server)
	notification, err := prov.VerifyNotification(
		context.Background(),
		`{"orderId":"sub2_order","paymentId":"pay_123","paymentType":"ALIPAY"}`,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, payment.NotificationStatusSuccess, notification.Status)
	require.Equal(t, "sub2_order", notification.OrderID)
	require.Equal(t, "pay_123", notification.TradeNo)
	require.Equal(t, 12.34, notification.Amount)
	require.Equal(t, "FACE_TO_FACE_PAYMENT", notification.Metadata["payment_channel"])
	require.Equal(t, testMuyinPlatform, notification.Metadata["platform"])
}

func TestMuyinVerifyNotificationRejectsNonSuccessQuery(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apis/platform.payment.muyin.site/v1alpha1/queryPaymentByPaymentId":
			_, _ = w.Write([]byte(`{
				"orderId":"sub2_order",
				"paymentId":"pay_123",
				"paymentType":"ALIPAY",
				"status":"UNPAID",
				"amount":12.34
			}`))
		case "/apis/platform.payment.muyin.site/v1alpha1/listPaymentInfo":
			_, _ = w.Write([]byte(`{"items":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	prov := mustTestMuyinProvider(t, server)
	_, err := prov.VerifyNotification(
		context.Background(),
		`{"orderId":"sub2_order","paymentId":"pay_123","paymentType":"ALIPAY"}`,
		nil,
	)
	require.ErrorContains(t, err, "not SUCCESS")
}

func TestMuyinRefundUnsupportedAndCancelClosesPayment(t *testing.T) {
	t.Parallel()

	var closePayload muyinPaymentRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/apis/platform.payment.muyin.site/v1alpha1/closePayment", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&closePayload))
		_, _ = w.Write([]byte(`{"orderId":"sub2_order","paymentType":"ALIPAY","status":"FAILED","message":"closed"}`))
	}))
	defer server.Close()

	prov := mustTestMuyinProvider(t, server)
	_, err := prov.Refund(context.Background(), payment.RefundRequest{OrderID: "sub2_order"})
	require.ErrorContains(t, err, "not supported")
	require.NoError(t, prov.CancelPayment(context.Background(), "sub2_order"))
	require.Equal(t, "sub2_order", closePayload.OrderID)
	require.Equal(t, muyinAlipayType, closePayload.PaymentType)
}

func mustTestMuyinProvider(t *testing.T, server *httptest.Server) *Muyin {
	t.Helper()
	prov, err := NewMuyin("1", map[string]string{
		"token":     "tok",
		"apiBase":   server.URL,
		"notifyUrl": "https://merchant.example.com/api/v1/payment/webhook/muyin",
		"returnUrl": "https://merchant.example.com/payment/result",
		"platform":  testMuyinPlatform,
	})
	require.NoError(t, err)
	return prov
}
