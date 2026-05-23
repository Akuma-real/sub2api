package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestCreateClaudeTestPayload_SpeedModeUsesLongPrompt(t *testing.T) {
	payload, err := createTestPayload("claude-sonnet-4-5", AccountTestModeSpeed)
	require.NoError(t, err)

	body, err := json.Marshal(payload)
	require.NoError(t, err)
	require.Equal(t, float64(speedTestOutputTokens), gjson.GetBytes(body, "max_tokens").Float())
	require.Equal(t, speedTestPrompt, gjson.GetBytes(body, "messages.0.content.0.text").String())
	require.True(t, gjson.GetBytes(body, "stream").Bool())
}

func TestCreateOpenAIResponsesTestPayload_SpeedModeUsesLongPrompt(t *testing.T) {
	payload := createOpenAITestPayload("gpt-5.4", false, AccountTestModeSpeed)

	body, err := json.Marshal(payload)
	require.NoError(t, err)
	require.Equal(t, float64(speedTestOutputTokens), gjson.GetBytes(body, "max_output_tokens").Float())
	require.Equal(t, speedTestPrompt, gjson.GetBytes(body, "input.0.content.0.text").String())
	require.True(t, gjson.GetBytes(body, "stream").Bool())
}

func TestCreateOpenAIChatCompletionsTestPayload_SpeedModeUsesLongPrompt(t *testing.T) {
	payload := createOpenAIChatCompletionsTestPayload("gpt-5.4", "ignored prompt", AccountTestModeSpeed)

	body, err := json.Marshal(payload)
	require.NoError(t, err)
	require.Equal(t, float64(speedTestOutputTokens), gjson.GetBytes(body, "max_tokens").Float())
	require.Equal(t, speedTestPrompt, gjson.GetBytes(body, "messages.0.content").String())
	require.True(t, gjson.GetBytes(body, "stream").Bool())
}

func TestCreateGeminiTestPayload_SpeedModeUsesGenerationConfig(t *testing.T) {
	payload := createGeminiTestPayload("gemini-2.5-flash", "ignored prompt", AccountTestModeSpeed)

	require.Equal(t, float64(speedTestOutputTokens), gjson.GetBytes(payload, "generationConfig.maxOutputTokens").Float())
	require.Equal(t, speedTestPrompt, gjson.GetBytes(payload, "contents.0.parts.0.text").String())
}

func TestAccountTestMetricsCompletionEventComputesGenerationThroughput(t *testing.T) {
	start := time.Now().Add(-2 * time.Second)
	metrics := &accountTestMetrics{
		startTime:      start,
		firstTokenTime: start.Add(500 * time.Millisecond),
		outputTokens:   750,
		outputChars:    3000,
	}

	event := metrics.CompletionEvent(true, "")

	require.True(t, event.DurationMs >= 1900)
	require.InDelta(t, 1500, event.GenerationMs, 120)
	require.InDelta(t, 500, event.OutputTokensPerSec, 60)
	require.InDelta(t, 2000, event.OutputCharsPerSec, 240)
}

func TestAccountTestService_OpenAIAPIKeySpeedModeSendsResponsesPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	account := Account{
		ID:          15,
		Name:        "openai-apikey",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "sk-test",
			"base_url": "https://example.com/v1",
		},
	}
	repo := &snapshotUpdateAccountRepo{
		stubOpenAIAccountRepo: stubOpenAIAccountRepo{accounts: []Account{account}},
	}
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(bytes.NewBufferString(
			"data: {\"type\":\"response.output_text.delta\",\"delta\":\"hello\"}\n\n" +
				"data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":20,\"output_tokens\":100}}}\n\n",
		)),
	}}
	svc := &AccountTestService{
		accountRepo:  repo,
		httpUpstream: upstream,
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/15/test", bytes.NewReader(nil))

	err := svc.TestAccountConnection(c, account.ID, "gpt-5.4", "", AccountTestModeSpeed)
	require.NoError(t, err)

	require.Equal(t, "https://example.com/v1/responses", upstream.lastReq.URL.String())
	require.Equal(t, float64(speedTestOutputTokens), gjson.GetBytes(upstream.lastBody, "max_output_tokens").Float())
	require.Equal(t, speedTestPrompt, gjson.GetBytes(upstream.lastBody, "input.0.content.0.text").String())
	require.Contains(t, rec.Body.String(), `"mode":"speed"`)
	require.Contains(t, rec.Body.String(), `"output_tokens":100`)
}
