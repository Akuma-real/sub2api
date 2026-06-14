package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	OpenAIDualAttemptRolePrimary   = "primary"
	OpenAIDualAttemptRoleSecondary = "secondary"

	OpenAIDualAttemptOutcomePending = "pending"
	OpenAIDualAttemptOutcomeWinner  = "winner"
	OpenAIDualAttemptOutcomeLoser   = "loser"
	OpenAIDualAttemptOutcomeSkipped = "skipped"

	OpenAIDualAttemptStatusCreated    = "created"
	OpenAIDualAttemptStatusDispatched = "dispatched"
	OpenAIDualAttemptStatusCompleted  = "completed"
	OpenAIDualAttemptStatusCanceled   = "canceled"
	OpenAIDualAttemptStatusFailed     = "failed"

	OpenAIDualBillingBasisTerminalUsage = "terminal_usage"
	OpenAIDualBillingBasisPartialUsage  = "partial_usage"
	OpenAIDualBillingBasisNoUsage       = "dispatched_no_usage"
	OpenAIDualBillingBasisNotDispatched = "not_dispatched"
	OpenAIDualBillingBasisUnsupportedWS = "dual_unsupported_ws_adapter"
)

const (
	OpenAIDualUnsupportedStreamingAdapter = "dual_unsupported_streaming_adapter"
)

type OpenAIDualAttemptPlan struct {
	Enabled          bool
	RequestID        string
	FirstTimeout     time.Duration
	StartedAt        time.Time
	PrimaryStarted   bool
	SecondaryStarted bool
}

func NewOpenAIDualAttemptPlan(apiKey *APIKey, requestID string, now time.Time) *OpenAIDualAttemptPlan {
	if !EffectiveOpenAIDualProtectionEnabled(apiKey) {
		return &OpenAIDualAttemptPlan{}
	}
	if now.IsZero() {
		now = time.Now()
	}
	timeoutMS := OpenAIDualFirstResponseTimeout(apiKey)
	return &OpenAIDualAttemptPlan{
		Enabled:      true,
		RequestID:    strings.TrimSpace(requestID),
		FirstTimeout: time.Duration(timeoutMS) * time.Millisecond,
		StartedAt:    now,
	}
}

func OpenAIDualProtectionSupportedForRequest(apiKey *APIKey, stream bool, endpoint string) (bool, string) {
	if !EffectiveOpenAIDualProtectionEnabled(apiKey) {
		return false, ""
	}
	endpoint = strings.ToLower(strings.TrimSpace(endpoint))
	if strings.Contains(endpoint, "ws") || strings.Contains(endpoint, "realtime") {
		return false, OpenAIDualBillingBasisUnsupportedWS
	}
	if stream {
		return false, OpenAIDualUnsupportedStreamingAdapter
	}
	return true, ""
}

func (p *OpenAIDualAttemptPlan) NextRole(now time.Time) string {
	if p == nil || !p.Enabled {
		return OpenAIDualAttemptRolePrimary
	}
	if !p.PrimaryStarted {
		p.PrimaryStarted = true
		return OpenAIDualAttemptRolePrimary
	}
	if !p.SecondaryStarted {
		if now.IsZero() {
			now = time.Now()
		}
		if p.FirstTimeout <= 0 || now.Sub(p.StartedAt) >= p.FirstTimeout {
			p.SecondaryStarted = true
			return OpenAIDualAttemptRoleSecondary
		}
	}
	return OpenAIDualAttemptRolePrimary
}

func (p *OpenAIDualAttemptPlan) ResultForSuccess(attempts []OpenAIDualAttempt, winnerAttemptID string) *OpenAIDualProtectionResult {
	if p == nil || !p.Enabled || len(attempts) == 0 {
		return nil
	}
	winnerAttemptID = strings.TrimSpace(winnerAttemptID)
	normalizedWinnerRole := ""
	switch strings.ToLower(winnerAttemptID) {
	case OpenAIDualAttemptRolePrimary, OpenAIDualAttemptRoleSecondary:
		normalizedWinnerRole = normalizeOpenAIDualAttemptRole(winnerAttemptID)
	}
	out := make([]OpenAIDualAttempt, 0, len(attempts))
	extraCost := 0.0
	for _, attempt := range attempts {
		attempt.Normalize()
		isWinner := false
		if winnerAttemptID != "" && attempt.AttemptID == winnerAttemptID {
			isWinner = true
		} else if normalizedWinnerRole != "" && attempt.Role == normalizedWinnerRole {
			isWinner = true
		}
		if isWinner {
			attempt.Outcome = OpenAIDualAttemptOutcomeWinner
		} else if attempt.Status == OpenAIDualAttemptStatusDispatched || attempt.Status == OpenAIDualAttemptStatusCompleted || attempt.Status == OpenAIDualAttemptStatusFailed || attempt.Status == OpenAIDualAttemptStatusCanceled {
			attempt.Outcome = OpenAIDualAttemptOutcomeLoser
			extraCost += attempt.BilledCost
		}
		out = append(out, attempt)
	}
	return &OpenAIDualProtectionResult{
		Enabled:      true,
		AttemptCount: len(out),
		ExtraCost:    extraCost,
		Attempts:     out,
	}
}

func (p *OpenAIDualAttemptPlan) ResultForUnsupported(endpoint, method, reason string) *OpenAIDualProtectionResult {
	if p == nil || !p.Enabled {
		return nil
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = OpenAIDualBillingBasisUnsupportedWS
	}
	basis := OpenAIDualBillingBasisNotDispatched
	attempt := OpenAIDualAttempt{
		RequestID:    p.RequestID,
		AttemptID:    OpenAIDualAttemptRoleSecondary,
		Endpoint:     strings.TrimSpace(endpoint),
		Method:       strings.TrimSpace(method),
		Role:         OpenAIDualAttemptRoleSecondary,
		Outcome:      OpenAIDualAttemptOutcomeSkipped,
		Status:       OpenAIDualAttemptStatusCreated,
		BillingBasis: &basis,
		CancelReason: &reason,
		Metadata:     map[string]any{"unsupported_reason": reason},
	}
	attempt.Normalize()
	return &OpenAIDualProtectionResult{
		Enabled:      true,
		AttemptCount: 1,
		ExtraCost:    0,
		Attempts:     []OpenAIDualAttempt{attempt},
	}
}

func NewOpenAIDualUnsupportedResult(apiKey *APIKey, requestID, endpoint, method, reason string) *OpenAIDualProtectionResult {
	plan := NewOpenAIDualAttemptPlan(apiKey, requestID, time.Now())
	return plan.ResultForUnsupported(endpoint, method, reason)
}

type CapturedOpenAIResponse struct {
	Status int
	Header http.Header
	Body   []byte
}

func (r *CapturedOpenAIResponse) Replay(c *gin.Context) {
	if r == nil || c == nil || c.Writer == nil || c.Writer.Written() {
		return
	}
	for key, values := range r.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
	status := r.Status
	if status <= 0 {
		status = http.StatusOK
	}
	c.Writer.WriteHeader(status)
	c.Writer.WriteHeaderNow()
	if len(r.Body) > 0 {
		_, _ = c.Writer.Write(r.Body)
	}
}

func NewOpenAIDualCaptureContext(c *gin.Context) (*gin.Context, *OpenAIDualCaptureWriter) {
	captured := &OpenAIDualCaptureWriter{
		header: make(http.Header),
		status: http.StatusOK,
		size:   -1,
	}
	if c == nil {
		return nil, captured
	}
	cp := c.Copy()
	cp.Writer = captured
	return cp, captured
}

type OpenAIDualCaptureWriter struct {
	header http.Header
	body   bytes.Buffer
	status int
	size   int
}

func (w *OpenAIDualCaptureWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *OpenAIDualCaptureWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	n, err := w.body.Write(data)
	if n > 0 {
		w.size += n
	}
	return n, err
}

func (w *OpenAIDualCaptureWriter) WriteString(data string) (int, error) {
	return w.Write([]byte(data))
}

func (w *OpenAIDualCaptureWriter) WriteHeader(code int) {
	if code <= 0 {
		return
	}
	if w.Written() {
		return
	}
	w.status = code
	w.size = 0
}

func (w *OpenAIDualCaptureWriter) WriteHeaderNow() {
	if !w.Written() {
		if w.status <= 0 {
			w.status = http.StatusOK
		}
		w.size = 0
	}
}

func (w *OpenAIDualCaptureWriter) Status() int {
	if w.status <= 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *OpenAIDualCaptureWriter) Size() int {
	return w.size
}

func (w *OpenAIDualCaptureWriter) Written() bool {
	return w.size >= 0
}

func (w *OpenAIDualCaptureWriter) Flush() {
	w.WriteHeaderNow()
}

func (w *OpenAIDualCaptureWriter) Pusher() http.Pusher {
	return nil
}

func (w *OpenAIDualCaptureWriter) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}

func (w *OpenAIDualCaptureWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("openai dual capture writer does not support hijack")
}

func (w *OpenAIDualCaptureWriter) CapturedResponse() *CapturedOpenAIResponse {
	if w == nil {
		return nil
	}
	return &CapturedOpenAIResponse{
		Status: w.Status(),
		Header: w.Header().Clone(),
		Body:   append([]byte(nil), w.body.Bytes()...),
	}
}

type OpenAIDualAttempt struct {
	ID                   int64
	RequestID            string
	AttemptID            string
	APIKeyID             int64
	UserID               int64
	AccountID            *int64
	Endpoint             string
	Method               string
	Role                 string
	Outcome              string
	ServiceTier          *string
	Status               string
	BillingBasis         *string
	EstimatedCost        float64
	ActualCost           float64
	BilledCost           float64
	UpstreamDispatchedAt *time.Time
	CancelReason         *string
	Metadata             map[string]any
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (a *OpenAIDualAttempt) Normalize() {
	if a == nil {
		return
	}
	a.RequestID = strings.TrimSpace(a.RequestID)
	a.AttemptID = strings.TrimSpace(a.AttemptID)
	if a.AttemptID == "" {
		a.AttemptID = OpenAIDualAttemptRolePrimary
	}
	a.Endpoint = strings.TrimSpace(a.Endpoint)
	a.Method = strings.ToUpper(strings.TrimSpace(a.Method))
	if a.Method == "" {
		a.Method = "POST"
	}
	a.Role = normalizeOpenAIDualAttemptRole(a.Role)
	a.Outcome = normalizeOpenAIDualAttemptOutcome(a.Outcome)
	a.Status = normalizeOpenAIDualAttemptStatus(a.Status)
	if a.Metadata == nil {
		a.Metadata = map[string]any{}
	}
}

func normalizeOpenAIDualAttemptRole(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case OpenAIDualAttemptRoleSecondary:
		return OpenAIDualAttemptRoleSecondary
	default:
		return OpenAIDualAttemptRolePrimary
	}
}

func normalizeOpenAIDualAttemptOutcome(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case OpenAIDualAttemptOutcomeWinner:
		return OpenAIDualAttemptOutcomeWinner
	case OpenAIDualAttemptOutcomeLoser:
		return OpenAIDualAttemptOutcomeLoser
	case OpenAIDualAttemptOutcomeSkipped:
		return OpenAIDualAttemptOutcomeSkipped
	default:
		return OpenAIDualAttemptOutcomePending
	}
}

func normalizeOpenAIDualAttemptStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case OpenAIDualAttemptStatusDispatched:
		return OpenAIDualAttemptStatusDispatched
	case OpenAIDualAttemptStatusCompleted:
		return OpenAIDualAttemptStatusCompleted
	case OpenAIDualAttemptStatusCanceled:
		return OpenAIDualAttemptStatusCanceled
	case OpenAIDualAttemptStatusFailed:
		return OpenAIDualAttemptStatusFailed
	default:
		return OpenAIDualAttemptStatusCreated
	}
}

type OpenAIDualAttemptRepository interface {
	Upsert(ctx context.Context, attempt *OpenAIDualAttempt) error
	ListByRequest(ctx context.Context, requestID string, apiKeyID int64) ([]OpenAIDualAttempt, error)
}

func (r *OpenAIForwardResult) DualProtectionAttempts(requestID string, apiKey *APIKey, user *User, account *Account, usageLog *UsageLog) []OpenAIDualAttempt {
	if r == nil || apiKey == nil || user == nil {
		return nil
	}
	if r.DualProtection != nil && len(r.DualProtection.Attempts) > 0 {
		attempts := make([]OpenAIDualAttempt, 0, len(r.DualProtection.Attempts))
		for _, attempt := range r.DualProtection.Attempts {
			fillOpenAIDualAttemptDefaults(&attempt, requestID, apiKey, user, account, usageLog)
			attempts = append(attempts, attempt)
		}
		return attempts
	}
	return nil
}

func fillOpenAIDualAttemptDefaults(attempt *OpenAIDualAttempt, requestID string, apiKey *APIKey, user *User, account *Account, usageLog *UsageLog) {
	if attempt == nil {
		return
	}
	if attempt.RequestID == "" {
		attempt.RequestID = requestID
	}
	if attempt.APIKeyID == 0 && apiKey != nil {
		attempt.APIKeyID = apiKey.ID
	}
	if attempt.UserID == 0 && user != nil {
		attempt.UserID = user.ID
	}
	if attempt.AccountID == nil && account != nil {
		accountID := account.ID
		attempt.AccountID = &accountID
	}
	if attempt.Endpoint == "" && usageLog != nil {
		if usageLog.UpstreamEndpoint != nil && strings.TrimSpace(*usageLog.UpstreamEndpoint) != "" {
			attempt.Endpoint = strings.TrimSpace(*usageLog.UpstreamEndpoint)
		} else if usageLog.InboundEndpoint != nil {
			attempt.Endpoint = strings.TrimSpace(*usageLog.InboundEndpoint)
		}
	}
	if attempt.ServiceTier == nil && usageLog != nil && usageLog.ServiceTier != nil {
		attempt.ServiceTier = usageLog.ServiceTier
	}
	if attempt.Metadata == nil {
		attempt.Metadata = map[string]any{}
	}
	attempt.Normalize()
}
