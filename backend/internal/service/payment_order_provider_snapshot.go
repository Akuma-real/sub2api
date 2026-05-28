package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
)

type paymentOrderProviderSnapshot struct {
	SchemaVersion      int
	ProviderInstanceID string
	ProviderKey        string
	PaymentMode        string
	MerchantAppID      string
	MerchantID         string
	Currency           string
	PaymentChannel     string
	Platform           string
}

func psOrderProviderSnapshot(order *dbent.PaymentOrder) *paymentOrderProviderSnapshot {
	if order == nil || len(order.ProviderSnapshot) == 0 {
		return nil
	}

	snapshot := &paymentOrderProviderSnapshot{
		SchemaVersion:      psSnapshotIntValue(order.ProviderSnapshot["schema_version"]),
		ProviderInstanceID: psSnapshotStringValue(order.ProviderSnapshot["provider_instance_id"]),
		ProviderKey:        psSnapshotStringValue(order.ProviderSnapshot["provider_key"]),
		PaymentMode:        psSnapshotStringValue(order.ProviderSnapshot["payment_mode"]),
		MerchantAppID:      psSnapshotStringValue(order.ProviderSnapshot["merchant_app_id"]),
		MerchantID:         psSnapshotStringValue(order.ProviderSnapshot["merchant_id"]),
		Currency:           psSnapshotStringValue(order.ProviderSnapshot["currency"]),
		PaymentChannel:     psSnapshotStringValue(order.ProviderSnapshot["payment_channel"]),
		Platform:           psSnapshotStringValue(order.ProviderSnapshot["platform"]),
	}
	if snapshot.SchemaVersion == 0 &&
		snapshot.ProviderInstanceID == "" &&
		snapshot.ProviderKey == "" &&
		snapshot.PaymentMode == "" &&
		snapshot.MerchantAppID == "" &&
		snapshot.MerchantID == "" &&
		snapshot.Currency == "" &&
		snapshot.PaymentChannel == "" &&
		snapshot.Platform == "" {
		return nil
	}
	return snapshot
}

func psSnapshotStringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}

func psSnapshotIntValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return n
		}
	}
	return 0
}

func (s *PaymentService) resolveSnapshotOrderProviderInstance(ctx context.Context, order *dbent.PaymentOrder, snapshot *paymentOrderProviderSnapshot) (*dbent.PaymentProviderInstance, error) {
	if s == nil || s.entClient == nil || order == nil || snapshot == nil {
		return nil, nil
	}

	snapshotInstanceID := strings.TrimSpace(snapshot.ProviderInstanceID)
	columnInstanceID := strings.TrimSpace(psStringValue(order.ProviderInstanceID))
	if snapshotInstanceID == "" {
		snapshotInstanceID = columnInstanceID
	}
	if snapshotInstanceID == "" {
		return nil, fmt.Errorf("order %d provider snapshot is missing provider_instance_id", order.ID)
	}
	if columnInstanceID != "" && snapshot.ProviderInstanceID != "" && !strings.EqualFold(columnInstanceID, snapshot.ProviderInstanceID) {
		return nil, fmt.Errorf("order %d provider snapshot instance mismatch: snapshot=%s order=%s", order.ID, snapshot.ProviderInstanceID, columnInstanceID)
	}

	instID, err := strconv.ParseInt(snapshotInstanceID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("order %d provider snapshot instance id is invalid: %s", order.ID, snapshotInstanceID)
	}

	inst, err := s.entClient.PaymentProviderInstance.Get(ctx, instID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, fmt.Errorf("order %d provider snapshot instance %s is missing", order.ID, snapshotInstanceID)
		}
		return nil, err
	}

	if snapshot.ProviderKey != "" && !strings.EqualFold(strings.TrimSpace(inst.ProviderKey), snapshot.ProviderKey) {
		return nil, fmt.Errorf("order %d provider snapshot key mismatch: snapshot=%s instance=%s", order.ID, snapshot.ProviderKey, inst.ProviderKey)
	}

	return inst, nil
}

func expectedNotificationProviderKeyForOrder(registry *payment.Registry, order *dbent.PaymentOrder, instanceProviderKey string) string {
	if order == nil {
		return strings.TrimSpace(instanceProviderKey)
	}

	orderProviderKey := psStringValue(order.ProviderKey)
	if snapshot := psOrderProviderSnapshot(order); snapshot != nil && snapshot.ProviderKey != "" {
		orderProviderKey = snapshot.ProviderKey
	}

	return expectedNotificationProviderKey(registry, order.PaymentType, orderProviderKey, instanceProviderKey)
}

func validateProviderSnapshotMetadata(order *dbent.PaymentOrder, providerKey string, metadata map[string]string) error {
	if order == nil || len(metadata) == 0 {
		return nil
	}

	snapshot := psOrderProviderSnapshot(order)
	if snapshot == nil {
		return nil
	}

	switch strings.TrimSpace(providerKey) {
	case payment.TypeWxpay:
		if expected := strings.TrimSpace(snapshot.MerchantAppID); expected != "" {
			actual := strings.TrimSpace(metadata["appid"])
			if actual == "" {
				return fmt.Errorf("wxpay notification missing appid")
			}
			if !strings.EqualFold(expected, actual) {
				return fmt.Errorf("wxpay appid mismatch: expected %s, got %s", expected, actual)
			}
		}
		if expected := strings.TrimSpace(snapshot.MerchantID); expected != "" {
			actual := strings.TrimSpace(metadata["mchid"])
			if actual == "" {
				return fmt.Errorf("wxpay notification missing mchid")
			}
			if !strings.EqualFold(expected, actual) {
				return fmt.Errorf("wxpay mchid mismatch: expected %s, got %s", expected, actual)
			}
		}
		if expected := strings.TrimSpace(snapshot.Currency); expected != "" {
			actual := strings.ToUpper(strings.TrimSpace(metadata["currency"]))
			if actual == "" {
				return fmt.Errorf("wxpay notification missing currency")
			}
			if !strings.EqualFold(expected, actual) {
				return fmt.Errorf("wxpay currency mismatch: expected %s, got %s", expected, actual)
			}
		}
		if actual := strings.TrimSpace(metadata["trade_state"]); actual != "" && !strings.EqualFold(actual, "SUCCESS") {
			return fmt.Errorf("wxpay trade_state mismatch: expected SUCCESS, got %s", actual)
		}
	case payment.TypeAlipay:
		if expected := strings.TrimSpace(snapshot.MerchantAppID); expected != "" {
			actual := strings.TrimSpace(metadata["app_id"])
			if actual == "" {
				return fmt.Errorf("alipay app_id missing")
			}
			if !strings.EqualFold(expected, actual) {
				return fmt.Errorf("alipay app_id mismatch: expected %s, got %s", expected, actual)
			}
		}
	case payment.TypeEasyPay:
		if expected := strings.TrimSpace(snapshot.MerchantID); expected != "" {
			actual := strings.TrimSpace(metadata["pid"])
			if actual == "" {
				return fmt.Errorf("easypay pid missing")
			}
			if !strings.EqualFold(expected, actual) {
				return fmt.Errorf("easypay pid mismatch: expected %s, got %s", expected, actual)
			}
		}
	case payment.TypeMuyin:
		if expected := strings.TrimSpace(order.PaymentTradeNo); expected != "" && !strings.EqualFold(expected, strings.TrimSpace(order.OutTradeNo)) {
			actual := strings.TrimSpace(metadata["payment_id"])
			if actual == "" {
				return fmt.Errorf("muyin payment_id missing")
			}
			if !strings.EqualFold(expected, actual) {
				return fmt.Errorf("muyin payment_id mismatch: expected %s, got %s", expected, actual)
			}
		}
		if expected := muyinExpectedPaymentType(order.PaymentType); expected != "" {
			actual := strings.TrimSpace(metadata["payment_type"])
			if actual == "" {
				return fmt.Errorf("muyin payment_type missing")
			}
			if !strings.EqualFold(expected, actual) {
				return fmt.Errorf("muyin payment_type mismatch: expected %s, got %s", expected, actual)
			}
		}
		if expected := strings.TrimSpace(snapshot.PaymentChannel); expected != "" {
			actual := strings.TrimSpace(metadata["payment_channel"])
			if actual != "" && !strings.EqualFold(expected, actual) {
				return fmt.Errorf("muyin payment_channel mismatch: expected %s, got %s", expected, actual)
			}
		}
		if expected := strings.TrimSpace(snapshot.Platform); expected != "" {
			actual := strings.TrimSpace(metadata["platform"])
			if actual != "" && !strings.EqualFold(expected, actual) {
				return fmt.Errorf("muyin platform mismatch: expected %s, got %s", expected, actual)
			}
		}
		if actual := strings.TrimSpace(metadata["status"]); actual != "" && !strings.EqualFold(actual, "SUCCESS") {
			return fmt.Errorf("muyin status mismatch: expected SUCCESS, got %s", actual)
		}
	case payment.TypeStripe:
		if expected := strings.TrimSpace(snapshot.Currency); expected != "" {
			actual := strings.ToUpper(strings.TrimSpace(metadata["currency"]))
			if actual == "" {
				return fmt.Errorf("stripe notification missing currency")
			}
			if !strings.EqualFold(expected, actual) {
				return fmt.Errorf("stripe currency mismatch: expected %s, got %s", expected, actual)
			}
		}
	case payment.TypeAirwallex:
		if expected := strings.TrimSpace(snapshot.MerchantID); expected != "" {
			actual := strings.TrimSpace(metadata["account_id"])
			if actual == "" {
				return fmt.Errorf("airwallex account_id missing")
			}
			if !strings.EqualFold(expected, actual) {
				return fmt.Errorf("airwallex account_id mismatch: expected %s, got %s", expected, actual)
			}
		}
		if expected := strings.TrimSpace(snapshot.Currency); expected != "" {
			actual := strings.ToUpper(strings.TrimSpace(metadata["currency"]))
			if actual == "" {
				return fmt.Errorf("airwallex notification missing currency")
			}
			if !strings.EqualFold(expected, actual) {
				return fmt.Errorf("airwallex currency mismatch: expected %s, got %s", expected, actual)
			}
		}
		if actual := strings.TrimSpace(metadata["status"]); actual != "" && !strings.EqualFold(actual, "SUCCEEDED") {
			return fmt.Errorf("airwallex status mismatch: expected SUCCEEDED, got %s", actual)
		}
	}

	return nil
}

func muyinExpectedPaymentType(paymentType string) string {
	switch payment.GetBasePaymentType(strings.TrimSpace(paymentType)) {
	case payment.TypeAlipay:
		return "ALIPAY"
	case payment.TypeWxpay:
		return "WECHATPAY"
	default:
		return ""
	}
}

func providerMerchantIdentityMetadata(prov payment.Provider) map[string]string {
	if prov == nil {
		return nil
	}
	reporter, ok := prov.(payment.MerchantIdentityProvider)
	if !ok {
		return nil
	}
	return reporter.MerchantIdentityMetadata()
}
