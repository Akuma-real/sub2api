package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

const (
	muyinDefaultAPIBase       = "https://auth.muyin.site"
	muyinHTTPTimeout          = 10 * time.Second
	maxMuyinResponseSize      = 1 << 20 // 1MB
	maxMuyinErrorSummary      = 512
	muyinAlipayType           = "ALIPAY"
	muyinWechatPayType        = "WECHATPAY"
	muyinAlipayDefaultChannel = "FACE_TO_FACE_PAYMENT"
	muyinWxpayDefaultChannel  = "WECHATPAY_H5"
	muyinWxpayJSAPIChannel    = "WECHATPAY_JSAPI"
	muyinStatusSuccess        = "SUCCESS"
	muyinStatusFailed         = "FAILED"
)

// Muyin implements payment.Provider for the merchant payment platform documented
// in docs/payment-api.md.
type Muyin struct {
	instanceID string
	config     map[string]string
	httpClient *http.Client
}

// NewMuyin creates a Muyin payment provider.
// config keys: token, apiBase, notifyUrl, returnUrl, platform, alipayChannel, wxpayChannel
func NewMuyin(instanceID string, config map[string]string) (*Muyin, error) {
	for _, k := range []string{"token", "notifyUrl", "returnUrl", "platform"} {
		if strings.TrimSpace(config[k]) == "" {
			return nil, fmt.Errorf("muyin config missing required key: %s", k)
		}
	}

	cfg := make(map[string]string, len(config)+4)
	for k, v := range config {
		cfg[k] = strings.TrimSpace(v)
	}
	if strings.TrimSpace(cfg["apiBase"]) == "" {
		cfg["apiBase"] = muyinDefaultAPIBase
	}
	cfg["apiBase"] = normalizeMuyinAPIBase(cfg["apiBase"])
	if strings.TrimSpace(cfg["alipayChannel"]) == "" {
		cfg["alipayChannel"] = muyinAlipayDefaultChannel
	}
	if strings.TrimSpace(cfg["wxpayChannel"]) == "" {
		cfg["wxpayChannel"] = muyinWxpayDefaultChannel
	}

	return &Muyin{
		instanceID: instanceID,
		config:     cfg,
		httpClient: &http.Client{Timeout: muyinHTTPTimeout},
	}, nil
}

func normalizeMuyinAPIBase(apiBase string) string {
	base := strings.TrimSpace(apiBase)
	if base == "" {
		return muyinDefaultAPIBase
	}
	if parsed, err := url.Parse(base); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		parsed.RawQuery = ""
		parsed.Fragment = ""
		parsed.RawPath = ""
		parsed.Path = trimMuyinEndpointPath(parsed.Path)
		return strings.TrimRight(parsed.String(), "/")
	}
	return strings.TrimRight(trimMuyinEndpointPath(base), "/")
}

func trimMuyinEndpointPath(path string) string {
	path = strings.TrimRight(strings.TrimSpace(path), "/")
	if idx := strings.Index(path, "/apis/"); idx >= 0 {
		return strings.TrimRight(path[:idx], "/")
	}
	return path
}

func (m *Muyin) apiBase() string {
	if m == nil {
		return ""
	}
	return normalizeMuyinAPIBase(m.config["apiBase"])
}

func (m *Muyin) Name() string        { return "MuYin" }
func (m *Muyin) ProviderKey() string { return payment.TypeMuyin }
func (m *Muyin) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeAlipay, payment.TypeWxpay}
}

func (m *Muyin) SupportedPaymentChannels(ctx context.Context, paymentType string) ([]MuyinPaymentChannelOption, error) {
	upstreamType := normalizeMuyinUpstreamType(paymentType)
	if upstreamType == "" {
		var err error
		upstreamType, err = muyinPaymentType(paymentType)
		if err != nil {
			return nil, err
		}
	}
	payload := muyinPaymentRequest{
		PaymentType: upstreamType,
		Platform:    strings.TrimSpace(m.config["platform"]),
	}
	body, status, err := m.postJSON(ctx, "/apis/platform.payment.muyin.site/v1alpha1/supportedPaymentChannels", payload)
	if err != nil {
		return nil, fmt.Errorf("muyin supported channels: %w", err)
	}
	channels, err := decodeMuyinPaymentChannelOptions(body)
	if err != nil {
		return nil, fmt.Errorf("parse supported channels (HTTP %d): %w", status, err)
	}
	return channels, nil
}

func (m *Muyin) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	amount, err := strconv.ParseFloat(strings.TrimSpace(req.Amount), 64)
	if err != nil || amount <= 0 {
		return nil, fmt.Errorf("muyin create: invalid amount %q", req.Amount)
	}
	upstreamType, err := muyinPaymentType(req.PaymentType)
	if err != nil {
		return nil, err
	}
	notifyURL, returnURL := m.resolveURLs(req)
	channel, err := m.paymentChannel(upstreamType, req)
	if err != nil {
		return nil, err
	}

	payload := muyinPaymentRequest{
		Subject:        strings.TrimSpace(req.Subject),
		Amount:         amount,
		PaymentType:    upstreamType,
		PaymentChannel: channel,
		OrderID:        strings.TrimSpace(req.OrderID),
		NotifyURL:      notifyURL,
		ReturnURL:      returnURL,
		Body:           strings.TrimSpace(req.Subject),
		ClientIP:       strings.TrimSpace(req.ClientIP),
		Platform:       strings.TrimSpace(m.config["platform"]),
		AdditionalData: m.additionalData(upstreamType, channel, req),
	}
	if enabled, ok := parseOptionalBool(m.config["enabledSucceedShow"]); ok {
		payload.EnabledSucceedShow = &enabled
	}

	result, err := m.postPayment(ctx, "/apis/platform.payment.muyin.site/v1alpha1/initiatePayment", payload)
	if err != nil {
		return nil, fmt.Errorf("muyin create: %w", err)
	}
	if strings.EqualFold(strings.TrimSpace(result.Status), muyinStatusFailed) {
		return nil, fmt.Errorf("muyin create failed: %s", result.message())
	}

	payURL := strings.TrimSpace(result.PaymentURL)
	qrCode := strings.TrimSpace(result.QRCode.QRCodeURL)
	if qrCode == "" {
		qrCode = strings.TrimSpace(result.QRCode.QRCodeContent)
	}
	if qrCode == "" {
		qrCode = strings.TrimSpace(result.QRCodeContent)
	}
	tradeNo := firstNonEmpty(result.PaymentID, result.TradeNo, result.OrderID)
	if payURL == "" && qrCode == "" && tradeNo == "" {
		return nil, fmt.Errorf("muyin create returned no payment result: %s", result.message())
	}
	return &payment.CreatePaymentResponse{
		TradeNo: tradeNo,
		PayURL:  payURL,
		QRCode:  qrCode,
	}, nil
}

func (m *Muyin) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	ref := strings.TrimSpace(tradeNo)
	if ref == "" {
		return nil, fmt.Errorf("muyin query: missing order identifier")
	}

	upstreamType := muyinPaymentTypeFromContext(ctx)
	result, err := m.queryPaymentByOrderID(ctx, ref, upstreamType)
	queryByPaymentID := false
	if err != nil {
		if paymentResult, paymentErr := m.queryPaymentByPaymentID(ctx, ref, upstreamType); paymentErr == nil {
			result = paymentResult
			err = nil
			queryByPaymentID = true
		}
	}
	if err != nil {
		return nil, fmt.Errorf("muyin query: %w", err)
	}
	metadata := result.metadata()
	var info *muyinPaymentInfo
	if info = m.lookupPaymentInfo(ctx, result, ref, queryByPaymentID); info != nil {
		mergeMuyinMetadata(metadata, info.metadata())
	}
	resp := result.toQueryOrderResponse(ref, metadata)
	resp.Amount = resolveMuyinAmount(result, info)
	return resp, nil
}

func (m *Muyin) VerifyNotification(ctx context.Context, rawBody string, _ map[string]string) (*payment.PaymentNotification, error) {
	fields := parseMuyinNotificationFields(rawBody)
	orderID := muyinField(fields, "orderId", "order_id", "out_trade_no", "merchantOrderId", "merchant_order_id")
	paymentID := muyinField(fields, "paymentId", "payment_id")
	upstreamType := normalizeMuyinUpstreamType(muyinField(fields, "paymentType", "payment_type"))
	if orderID == "" && paymentID == "" {
		return nil, fmt.Errorf("muyin notify: missing orderId/paymentId")
	}

	var result *muyinPaymentResult
	var err error
	if paymentID != "" {
		result, err = m.queryPaymentByPaymentID(ctx, paymentID, upstreamType)
	}
	if err != nil || result == nil {
		if orderID == "" {
			return nil, fmt.Errorf("muyin notify query: %w", err)
		}
		result, err = m.queryPaymentByOrderID(ctx, orderID, upstreamType)
	}
	if err != nil {
		return nil, fmt.Errorf("muyin notify query: %w", err)
	}

	confirmedOrderID := firstNonEmpty(result.OrderID, orderID)
	if confirmedOrderID == "" {
		return nil, fmt.Errorf("muyin notify query returned no orderId")
	}

	metadata := result.metadata()
	var info *muyinPaymentInfo
	if info = m.lookupPaymentInfo(ctx, result, firstNonEmpty(paymentID, orderID), paymentID != ""); info != nil {
		mergeMuyinMetadata(metadata, info.metadata())
	}
	status := strings.TrimSpace(metadata["status"])
	if status == "" {
		status = strings.TrimSpace(result.Status)
	}
	if !strings.EqualFold(status, muyinStatusSuccess) {
		return nil, fmt.Errorf("muyin notify query status is not SUCCESS: %s", firstNonEmpty(status, "empty"))
	}
	return &payment.PaymentNotification{
		TradeNo:  firstNonEmpty(result.PaymentID, result.TradeNo, paymentID, confirmedOrderID),
		OrderID:  confirmedOrderID,
		Amount:   resolveMuyinAmount(result, info),
		Status:   payment.NotificationStatusSuccess,
		RawData:  rawBody,
		Metadata: metadata,
	}, nil
}

func (m *Muyin) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("muyin refund is not supported")
}

func (m *Muyin) CancelPayment(ctx context.Context, tradeNo string) error {
	ref := strings.TrimSpace(tradeNo)
	if ref == "" {
		return fmt.Errorf("muyin close: missing order identifier")
	}
	upstreamType := muyinPaymentTypeFromContext(ctx)
	if _, err := m.closePayment(ctx, ref, "", upstreamType); err == nil {
		return nil
	}
	if _, err := m.closePayment(ctx, "", ref, upstreamType); err == nil {
		return nil
	} else {
		return fmt.Errorf("muyin close: %w", err)
	}
}

func (m *Muyin) resolveURLs(req payment.CreatePaymentRequest) (string, string) {
	notifyURL := strings.TrimSpace(req.NotifyURL)
	if notifyURL == "" {
		notifyURL = strings.TrimSpace(m.config["notifyUrl"])
	}
	returnURL := strings.TrimSpace(req.ReturnURL)
	if returnURL == "" {
		returnURL = strings.TrimSpace(m.config["returnUrl"])
	}
	return notifyURL, returnURL
}

func (m *Muyin) paymentChannel(upstreamType string, req payment.CreatePaymentRequest) (string, error) {
	switch upstreamType {
	case muyinAlipayType:
		return strings.TrimSpace(m.config["alipayChannel"]), nil
	case muyinWechatPayType:
		channel := strings.TrimSpace(m.config["wxpayChannel"])
		if channel == "" {
			channel = muyinWxpayDefaultChannel
		}
		if channel == muyinWxpayJSAPIChannel && strings.TrimSpace(req.OpenID) == "" {
			return "", fmt.Errorf("muyin wxpay jsapi requires openid")
		}
		if channel == muyinWxpayDefaultChannel && strings.TrimSpace(req.ClientIP) == "" {
			return "", fmt.Errorf("muyin wxpay h5 requires client ip")
		}
		return channel, nil
	default:
		return "", fmt.Errorf("unsupported muyin payment type: %s", upstreamType)
	}
}

func (m *Muyin) additionalData(upstreamType string, channel string, req payment.CreatePaymentRequest) *muyinAdditionalData {
	if upstreamType != muyinWechatPayType {
		return nil
	}
	data := &muyinAdditionalData{
		KeyValues:   map[string]string{},
		NameForKeys: map[string]string{},
	}
	if channel == muyinWxpayJSAPIChannel {
		if openID := strings.TrimSpace(req.OpenID); openID != "" {
			data.KeyValues["openid"] = openID
			data.NameForKeys["openid"] = "公众号用户 OpenID"
		}
	}
	if channel == muyinWxpayDefaultChannel {
		if clientIP := strings.TrimSpace(req.ClientIP); clientIP != "" {
			data.KeyValues["payerClientIp"] = clientIP
			data.NameForKeys["payerClientIp"] = "用户端 IP"
		}
		data.KeyValues["h5Type"] = "Wap"
		data.NameForKeys["h5Type"] = "场景类型"
		if appName := strings.TrimSpace(m.config["wxpayH5AppName"]); appName != "" {
			data.KeyValues["appName"] = appName
			data.NameForKeys["appName"] = "应用名"
		}
		if appURL := strings.TrimSpace(m.config["wxpayH5AppUrl"]); appURL != "" {
			data.KeyValues["appUrl"] = appURL
			data.NameForKeys["appUrl"] = "应用链接"
		}
	}
	if len(data.KeyValues) == 0 {
		return nil
	}
	return data
}

func muyinPaymentType(paymentType string) (string, error) {
	switch payment.GetBasePaymentType(strings.TrimSpace(paymentType)) {
	case payment.TypeAlipay:
		return muyinAlipayType, nil
	case payment.TypeWxpay:
		return muyinWechatPayType, nil
	default:
		return "", fmt.Errorf("muyin unsupported payment type: %s", paymentType)
	}
}

func normalizeMuyinUpstreamType(raw string) string {
	raw = strings.ToUpper(strings.TrimSpace(raw))
	switch raw {
	case "ALIPAY", "WECHATPAY":
		return raw
	default:
		return ""
	}
}

func muyinPaymentTypeFromContext(ctx context.Context) string {
	paymentType := strings.TrimSpace(payment.QueryPaymentTypeFromContext(ctx))
	if paymentType == "" {
		return ""
	}
	if upstreamType, err := muyinPaymentType(paymentType); err == nil {
		return upstreamType
	}
	return normalizeMuyinUpstreamType(paymentType)
}

func (m *Muyin) queryPaymentByOrderID(ctx context.Context, orderID string, upstreamType string) (*muyinPaymentResult, error) {
	return m.queryPayment(ctx, "/apis/platform.payment.muyin.site/v1alpha1/queryPaymentByOrderId", "orderId", orderID, upstreamType)
}

func (m *Muyin) queryPaymentByPaymentID(ctx context.Context, paymentID string, upstreamType string) (*muyinPaymentResult, error) {
	return m.queryPayment(ctx, "/apis/platform.payment.muyin.site/v1alpha1/queryPaymentByPaymentId", "paymentId", paymentID, upstreamType)
}

func (m *Muyin) closePayment(ctx context.Context, orderID string, paymentID string, upstreamType string) (*muyinPaymentResult, error) {
	payload := muyinPaymentRequest{
		OrderID:   strings.TrimSpace(orderID),
		PaymentID: strings.TrimSpace(paymentID),
		Platform:  strings.TrimSpace(m.config["platform"]),
	}
	return m.postPaymentWithTypeFallback(ctx, "/apis/platform.payment.muyin.site/v1alpha1/closePayment", payload, upstreamType)
}

func (m *Muyin) queryPayment(ctx context.Context, endpoint string, idField string, id string, upstreamType string) (*muyinPaymentResult, error) {
	payload := muyinPaymentRequest{
		Platform: strings.TrimSpace(m.config["platform"]),
	}
	switch idField {
	case "orderId":
		payload.OrderID = strings.TrimSpace(id)
	case "paymentId":
		payload.PaymentID = strings.TrimSpace(id)
	default:
		return nil, fmt.Errorf("unsupported query field: %s", idField)
	}
	return m.postPaymentWithTypeFallback(ctx, endpoint, payload, upstreamType)
}

func (m *Muyin) lookupPaymentInfo(ctx context.Context, result *muyinPaymentResult, fallbackRef string, fallbackRefIsPaymentID bool) *muyinPaymentInfo {
	orderID := strings.TrimSpace(result.OrderID)
	paymentID := strings.TrimSpace(result.PaymentID)
	if orderID == "" && !fallbackRefIsPaymentID {
		orderID = strings.TrimSpace(fallbackRef)
	}
	if paymentID == "" && fallbackRefIsPaymentID {
		paymentID = strings.TrimSpace(fallbackRef)
	}
	info, err := m.listPaymentInfo(ctx, orderID, paymentID)
	if err != nil {
		return nil
	}
	return info
}

func (m *Muyin) listPaymentInfo(ctx context.Context, orderID string, paymentID string) (*muyinPaymentInfo, error) {
	query := url.Values{}
	if orderID = strings.TrimSpace(orderID); orderID != "" {
		query.Set("orderId", orderID)
	}
	if paymentID = strings.TrimSpace(paymentID); paymentID != "" {
		query.Set("paymentId", paymentID)
	}
	if platform := strings.TrimSpace(m.config["platform"]); platform != "" {
		query.Set("platform", platform)
	}
	body, status, err := m.getJSON(ctx, "/apis/platform.payment.muyin.site/v1alpha1/listPaymentInfo", query)
	if err != nil {
		return nil, err
	}
	items, err := decodeMuyinPaymentInfoList(body)
	if err != nil {
		return nil, fmt.Errorf("parse payment info list (HTTP %d): %w", status, err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("muyin payment info not found")
	}
	return &items[0], nil
}

func (m *Muyin) postPaymentWithTypeFallback(ctx context.Context, endpoint string, payload muyinPaymentRequest, upstreamType string) (*muyinPaymentResult, error) {
	paymentTypes := []string{normalizeMuyinUpstreamType(upstreamType)}
	if paymentTypes[0] == "" {
		paymentTypes = []string{muyinAlipayType, muyinWechatPayType}
	}

	var firstErr error
	var lastResult *muyinPaymentResult
	for _, paymentType := range paymentTypes {
		payload.PaymentType = paymentType
		result, err := m.postPayment(ctx, endpoint, payload)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if len(paymentTypes) > 1 && muyinResultLooksMissing(result) {
			lastResult = result
			continue
		}
		return result, nil
	}
	if firstErr != nil {
		return nil, firstErr
	}
	if lastResult != nil {
		return lastResult, nil
	}
	return nil, fmt.Errorf("muyin request failed")
}

func (m *Muyin) postPayment(ctx context.Context, endpoint string, payload muyinPaymentRequest) (*muyinPaymentResult, error) {
	body, status, err := m.postJSON(ctx, endpoint, payload)
	if err != nil {
		return nil, err
	}
	result, err := decodeMuyinPaymentResult(body)
	if err != nil {
		return nil, fmt.Errorf("parse response (HTTP %d): %w", status, err)
	}
	return result, nil
}

func (m *Muyin) postJSON(ctx context.Context, endpoint string, payload any) ([]byte, int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.apiBase()+endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(m.config["token"]))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := m.httpClient
	if client == nil {
		client = &http.Client{Timeout: muyinHTTPTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxMuyinResponseSize))
	if readErr != nil {
		return nil, resp.StatusCode, readErr
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, summarizeMuyinResponse(body))
	}
	return body, resp.StatusCode, nil
}

func (m *Muyin) getJSON(ctx context.Context, endpoint string, query url.Values) ([]byte, int, error) {
	target := m.apiBase() + endpoint
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(m.config["token"]))
	req.Header.Set("Accept", "application/json")

	client := m.httpClient
	if client == nil {
		client = &http.Client{Timeout: muyinHTTPTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxMuyinResponseSize))
	if readErr != nil {
		return nil, resp.StatusCode, readErr
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, summarizeMuyinResponse(body))
	}
	return body, resp.StatusCode, nil
}

type muyinPaymentRequest struct {
	Subject            string               `json:"subject,omitempty"`
	Amount             float64              `json:"amount,omitempty"`
	PaymentType        string               `json:"paymentType,omitempty"`
	PaymentChannel     string               `json:"paymentChannel,omitempty"`
	OrderID            string               `json:"orderId,omitempty"`
	PaymentID          string               `json:"paymentId,omitempty"`
	NotifyURL          string               `json:"notifyUrl,omitempty"`
	ReturnURL          string               `json:"returnUrl,omitempty"`
	Body               string               `json:"body,omitempty"`
	ClientIP           string               `json:"clientIp,omitempty"`
	UserID             string               `json:"userId,omitempty"`
	Platform           string               `json:"platform,omitempty"`
	EnabledSucceedShow *bool                `json:"enabledSucceedShow,omitempty"`
	AdditionalData     *muyinAdditionalData `json:"additionalData,omitempty"`
}

type muyinAdditionalData struct {
	KeyValues   map[string]string `json:"keyValues,omitempty"`
	NameForKeys map[string]string `json:"nameForKeys,omitempty"`
}

type muyinQRCode struct {
	QRCodeContent string `json:"qrCodeContent"`
	QRCodeURL     string `json:"qrCodeUrl"`
}

type muyinPaymentResult struct {
	OrderID        string         `json:"orderId"`
	PaymentID      string         `json:"paymentId"`
	PaymentType    string         `json:"paymentType"`
	PaymentChannel string         `json:"paymentChannel"`
	PaymentURL     string         `json:"paymentUrl"`
	QRCode         muyinQRCode    `json:"qrCode"`
	QRCodeContent  string         `json:"qrCodeContent"`
	Status         string         `json:"status"`
	TradeNo        string         `json:"tradeNo"`
	Amount         float64        `json:"amount"`
	PaymentAmount  any            `json:"paymentAmount"`
	PaymentTime    string         `json:"paymentTime"`
	Message        string         `json:"message"`
	Response       *muyinResponse `json:"response"`
	AdditionalData map[string]any `json:"additionalData"`
}

type muyinPaymentInfo struct {
	OrderID            string `json:"orderId"`
	PaymentID          string `json:"paymentId"`
	PaymentAmount      any    `json:"paymentAmount"`
	PaymentType        string `json:"paymentType"`
	PaymentChannel     string `json:"paymentChannel"`
	PaymentStatus      string `json:"paymentStatus"`
	PaymentURL         string `json:"paymentUrl"`
	PaymentPageContent string `json:"paymentPageContent"`
	QRCodeContent      string `json:"qrCodeContent"`
	TradeNo            string `json:"tradeNo"`
	Platform           string `json:"platform"`
	Subject            string `json:"subject"`
	Body               string `json:"body"`
	ClientIP           string `json:"clientIp"`
	UserID             string `json:"userId"`
	PaymentAccount     string `json:"paymentAccount"`
	NotifyURL          string `json:"notifyUrl"`
	ReturnURL          string `json:"returnUrl"`
	CreateTime         string `json:"createTime"`
	PaymentTime        string `json:"paymentTime"`
}

type muyinResponse struct {
	Code any    `json:"code"`
	Msg  string `json:"msg"`
}

type MuyinPaymentChannelOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func (r *muyinPaymentResult) message() string {
	if r == nil {
		return ""
	}
	if msg := strings.TrimSpace(r.Message); msg != "" {
		return msg
	}
	if r.Response != nil {
		return strings.TrimSpace(r.Response.Msg)
	}
	return strings.TrimSpace(r.Status)
}

func (r *muyinPaymentResult) metadata() map[string]string {
	metadata := map[string]string{
		"payment_id":   strings.TrimSpace(r.PaymentID),
		"payment_type": strings.TrimSpace(r.PaymentType),
		"status":       strings.TrimSpace(r.Status),
	}
	if r.TradeNo != "" {
		metadata["trade_no"] = strings.TrimSpace(r.TradeNo)
	}
	if r.PaymentChannel != "" {
		metadata["payment_channel"] = strings.TrimSpace(r.PaymentChannel)
	}
	return metadata
}

func (r *muyinPaymentResult) amount() float64 {
	if r == nil {
		return 0
	}
	if r.Amount > 0 {
		return r.Amount
	}
	return parseMuyinAmount(r.PaymentAmount)
}

func (i *muyinPaymentInfo) amount() float64 {
	if i == nil {
		return 0
	}
	return parseMuyinAmount(i.PaymentAmount)
}

func (i *muyinPaymentInfo) metadata() map[string]string {
	if i == nil {
		return nil
	}
	metadata := map[string]string{}
	if i.PaymentID != "" {
		metadata["payment_id"] = strings.TrimSpace(i.PaymentID)
	}
	if i.PaymentType != "" {
		metadata["payment_type"] = strings.TrimSpace(i.PaymentType)
	}
	if i.PaymentStatus != "" {
		metadata["status"] = strings.TrimSpace(i.PaymentStatus)
	}
	if i.TradeNo != "" {
		metadata["trade_no"] = strings.TrimSpace(i.TradeNo)
	}
	if i.PaymentChannel != "" {
		metadata["payment_channel"] = strings.TrimSpace(i.PaymentChannel)
	}
	if i.Platform != "" {
		metadata["platform"] = strings.TrimSpace(i.Platform)
	}
	return metadata
}

func resolveMuyinAmount(result *muyinPaymentResult, info *muyinPaymentInfo) float64 {
	if amount := result.amount(); amount > 0 {
		return amount
	}
	return info.amount()
}

func parseMuyinAmount(raw any) float64 {
	switch value := raw.(type) {
	case nil:
		return 0
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case json.Number:
		parsed, err := value.Float64()
		if err == nil {
			return parsed
		}
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func mergeMuyinMetadata(dst map[string]string, src map[string]string) {
	for key, value := range src {
		if strings.TrimSpace(value) != "" {
			dst[key] = strings.TrimSpace(value)
		}
	}
}

func (r *muyinPaymentResult) toQueryOrderResponse(queryRef string, metadata map[string]string) *payment.QueryOrderResponse {
	status := payment.ProviderStatusPending
	switch strings.ToUpper(strings.TrimSpace(r.Status)) {
	case muyinStatusSuccess:
		status = payment.ProviderStatusPaid
	case muyinStatusFailed:
		status = payment.ProviderStatusFailed
	}
	if metadata == nil {
		metadata = r.metadata()
	}
	return &payment.QueryOrderResponse{
		TradeNo:  firstNonEmpty(r.PaymentID, r.TradeNo, r.OrderID, queryRef),
		Status:   status,
		Amount:   r.amount(),
		PaidAt:   strings.TrimSpace(r.PaymentTime),
		Metadata: metadata,
	}
}

func decodeMuyinPaymentResult(body []byte) (*muyinPaymentResult, error) {
	var wrapper struct {
		Data    json.RawMessage `json:"data"`
		Result  json.RawMessage `json:"result"`
		Message string          `json:"message"`
		Msg     string          `json:"msg"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, err
	}
	var firstNested *muyinPaymentResult
	for _, raw := range []json.RawMessage{wrapper.Data, wrapper.Result} {
		if len(raw) == 0 || string(raw) == "null" {
			continue
		}
		var nested muyinPaymentResult
		if err := json.Unmarshal(raw, &nested); err != nil {
			return nil, err
		}
		if nested.Message == "" {
			nested.Message = firstNonEmpty(wrapper.Message, wrapper.Msg)
		}
		if firstNested == nil {
			copied := nested
			firstNested = &copied
		}
		if hasMeaningfulMuyinResult(&nested) {
			return &nested, nil
		}
	}

	var direct muyinPaymentResult
	if err := json.Unmarshal(body, &direct); err != nil {
		return nil, err
	}
	if hasMeaningfulMuyinResult(&direct) || firstNested == nil {
		return &direct, nil
	}
	return firstNested, nil
}

func decodeMuyinPaymentChannelOptions(body []byte) ([]MuyinPaymentChannelOption, error) {
	var direct []MuyinPaymentChannelOption
	if err := json.Unmarshal(body, &direct); err == nil {
		return direct, nil
	}
	var wrapper struct {
		Data   []MuyinPaymentChannelOption `json:"data"`
		Result []MuyinPaymentChannelOption `json:"result"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, err
	}
	if wrapper.Data != nil {
		return wrapper.Data, nil
	}
	return wrapper.Result, nil
}

func decodeMuyinPaymentInfoList(body []byte) ([]muyinPaymentInfo, error) {
	var direct []muyinPaymentInfo
	if err := json.Unmarshal(body, &direct); err == nil {
		return direct, nil
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	for _, key := range []string{"items", "data", "result"} {
		if items, ok, err := decodeMuyinPaymentInfoListValue(root[key]); ok || err != nil {
			return items, err
		}
	}
	return nil, nil
}

func decodeMuyinPaymentInfoListValue(raw json.RawMessage) ([]muyinPaymentInfo, bool, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, false, nil
	}

	var direct []muyinPaymentInfo
	if err := json.Unmarshal(raw, &direct); err == nil {
		return direct, true, nil
	}

	var nested map[string]json.RawMessage
	if err := json.Unmarshal(raw, &nested); err != nil {
		return nil, false, err
	}
	for _, key := range []string{"items", "data", "result"} {
		if items, ok, err := decodeMuyinPaymentInfoListValue(nested[key]); ok || err != nil {
			return items, ok, err
		}
	}
	return nil, false, nil
}

func hasMeaningfulMuyinResult(result *muyinPaymentResult) bool {
	if result == nil {
		return false
	}
	return firstNonEmpty(result.OrderID, result.PaymentID, result.PaymentURL, result.TradeNo, result.Status) != ""
}

func muyinResultLooksMissing(result *muyinPaymentResult) bool {
	if result == nil {
		return false
	}
	message := strings.ToLower(result.message())
	if strings.Contains(message, "not found") ||
		strings.Contains(message, "not exist") ||
		strings.Contains(message, "no record") ||
		strings.Contains(message, "不存在") ||
		strings.Contains(message, "未找到") {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(result.Status), "UNKNOWN") &&
		firstNonEmpty(result.OrderID, result.PaymentID, result.PaymentURL, result.TradeNo) == ""
}

func parseMuyinNotificationFields(rawBody string) map[string]string {
	fields := map[string]string{}
	values, err := url.ParseQuery(rawBody)
	if err == nil && len(values) > 0 {
		for k := range values {
			if v := strings.TrimSpace(values.Get(k)); v != "" {
				fields[strings.ToLower(strings.TrimSpace(k))] = v
			}
		}
	}

	var payload any
	if err := json.Unmarshal([]byte(rawBody), &payload); err == nil {
		collectMuyinJSONFields(payload, fields)
	}
	return fields
}

func collectMuyinJSONFields(value any, fields map[string]string) {
	switch typed := value.(type) {
	case map[string]any:
		for k, v := range typed {
			key := strings.ToLower(strings.TrimSpace(k))
			switch scalar := v.(type) {
			case string:
				if strings.TrimSpace(scalar) != "" {
					fields[key] = strings.TrimSpace(scalar)
				}
			case float64:
				fields[key] = strconv.FormatFloat(scalar, 'f', -1, 64)
			case bool:
				fields[key] = strconv.FormatBool(scalar)
			default:
				collectMuyinJSONFields(v, fields)
			}
		}
	case []any:
		for _, item := range typed {
			collectMuyinJSONFields(item, fields)
		}
	}
}

func muyinField(fields map[string]string, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(fields[strings.ToLower(key)]); v != "" {
			return v
		}
	}
	return ""
}

func parseOptionalBool(raw string) (bool, bool) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return false, false
	}
	switch raw {
	case "true", "1", "yes", "y", "on":
		return true, true
	case "false", "0", "no", "n", "off":
		return false, true
	default:
		return false, false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func summarizeMuyinResponse(body []byte) string {
	summary := strings.Join(strings.Fields(string(body)), " ")
	if summary == "" {
		return "<empty>"
	}
	if len(summary) > maxMuyinErrorSummary {
		return summary[:maxMuyinErrorSummary] + "..."
	}
	return summary
}
