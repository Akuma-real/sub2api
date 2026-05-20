package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/proxyurl"
	"github.com/Wei-Shaw/sub2api/internal/pkg/proxyutil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const maxDiscoverModelsResponseBytes = 2 << 20

var modelDiscoverySecretQueryPattern = regexp.MustCompile(`(?i)([?&](?:key|api[_-]?key|access[_-]?token|token|authorization|x-api-key)=)[^&\s"']+`)

type upstreamModelDiscoveryEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DiscoverUpstreamModelsRequest represents an admin-only model discovery probe.
type DiscoverUpstreamModelsRequest struct {
	Platform string `json:"platform" binding:"required"`
	Type     string `json:"type"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key" binding:"required"`
	ProxyID  *int64 `json:"proxy_id"`
}

// DiscoverUpstreamModelsResponse contains model IDs discovered from an upstream.
type DiscoverUpstreamModelsResponse struct {
	Models []string `json:"models"`
}

// DiscoverUpstreamModels fetches model IDs from a configured upstream through the backend.
// POST /api/v1/admin/accounts/discover-models
func (h *AccountHandler) DiscoverUpstreamModels(c *gin.Context) {
	var req DiscoverUpstreamModelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	apiKey := strings.TrimSpace(req.APIKey)
	if apiKey == "" {
		response.BadRequest(c, "API key is required")
		return
	}

	endpoint, headers, err := buildUpstreamModelsRequest(req, apiKey)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	client, err := h.newModelDiscoveryHTTPClient(c.Request.Context(), req.ProxyID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	models, err := requestUpstreamModelIDs(c.Request.Context(), client, endpoint, headers)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}

	response.Success(c, DiscoverUpstreamModelsResponse{Models: models})
}

func buildUpstreamModelsRequest(req DiscoverUpstreamModelsRequest, apiKey string) (string, map[string]string, error) {
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	headers := map[string]string{}

	switch platform {
	case service.PlatformOpenAI, service.PlatformAntigravity:
		base, err := normalizeModelDiscoveryBaseURL(req.BaseURL, "https://api.openai.com", "v1")
		if err != nil {
			return "", nil, err
		}
		headers["Authorization"] = "Bearer " + apiKey
		if platform == service.PlatformAntigravity {
			headers["x-api-key"] = apiKey
			headers["anthropic-version"] = "2023-06-01"
		}
		return base + "/models", headers, nil
	case service.PlatformAnthropic:
		base, err := normalizeModelDiscoveryBaseURL(req.BaseURL, "https://api.anthropic.com", "v1")
		if err != nil {
			return "", nil, err
		}
		headers["x-api-key"] = apiKey
		headers["anthropic-version"] = "2023-06-01"
		return base + "/models", headers, nil
	case service.PlatformGemini:
		base, err := normalizeModelDiscoveryBaseURL(req.BaseURL, "https://generativelanguage.googleapis.com", "v1beta")
		if err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("%s/models?key=%s", base, url.QueryEscape(apiKey)), headers, nil
	default:
		return "", nil, fmt.Errorf("unsupported platform: %s", req.Platform)
	}
}

func normalizeModelDiscoveryBaseURL(raw, fallback, version string) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(raw), "/")
	if base == "" {
		base = strings.TrimRight(strings.TrimSpace(fallback), "/")
	}
	if base == "" {
		return "", errors.New("base URL is required")
	}

	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid base URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("base URL must use http or https")
	}

	if strings.HasSuffix(parsed.Path, "/v1") ||
		strings.HasSuffix(parsed.Path, "/v1/models") ||
		strings.HasSuffix(parsed.Path, "/v1beta") ||
		strings.HasSuffix(parsed.Path, "/v1beta/models") {
		return base, nil
	}
	return base + "/" + version, nil
}

func (h *AccountHandler) newModelDiscoveryHTTPClient(ctx context.Context, proxyID *int64) (*http.Client, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	if proxyID == nil || *proxyID <= 0 {
		return client, nil
	}

	proxy, err := h.adminService.GetProxy(ctx, *proxyID)
	if err != nil {
		return nil, fmt.Errorf("failed to load proxy: %w", err)
	}
	proxyURL := strings.TrimSpace(proxy.URL())
	if proxyURL == "" {
		return client, nil
	}

	_, parsedProxy, err := proxyurl.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("unexpected default transport type %T", http.DefaultTransport)
	}
	transport := defaultTransport.Clone()
	transport.Proxy = nil
	if err := proxyutil.ConfigureTransportProxy(transport, parsedProxy); err != nil {
		return nil, fmt.Errorf("configure proxy: %w", err)
	}
	client.Transport = transport
	return client, nil
}

func requestUpstreamModelIDs(ctx context.Context, client *http.Client, endpoint string, headers map[string]string) ([]string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %s", sanitizeModelDiscoveryError(err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDiscoverModelsResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}
	if len(body) > maxDiscoverModelsResponseBytes {
		return nil, errors.New("upstream response is too large")
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}
		return nil, fmt.Errorf("upstream returned %d: %s", resp.StatusCode, sanitizeModelDiscoveryError(message))
	}

	models, err := extractUpstreamModelIDs(body)
	if err != nil {
		return nil, err
	}
	return models, nil
}

func sanitizeModelDiscoveryError(message string) string {
	return modelDiscoverySecretQueryPattern.ReplaceAllString(message, `${1}[REDACTED]`)
}

func extractUpstreamModelIDs(body []byte) ([]string, error) {
	var payload struct {
		Data   []upstreamModelDiscoveryEntry `json:"data"`
		Models []upstreamModelDiscoveryEntry `json:"models"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		var arrayPayload []upstreamModelDiscoveryEntry
		if arrayErr := json.Unmarshal(body, &arrayPayload); arrayErr != nil {
			return nil, fmt.Errorf("invalid upstream model response: %w", err)
		}
		return dedupeAndSortModelIDs(arrayPayload)
	}

	entries := append([]upstreamModelDiscoveryEntry(nil), payload.Data...)
	entries = append(entries, payload.Models...)
	if len(entries) == 0 {
		var arrayPayload []upstreamModelDiscoveryEntry
		if err := json.Unmarshal(body, &arrayPayload); err == nil {
			entries = arrayPayload
		}
	}
	return dedupeAndSortModelIDs(entries)
}

func dedupeAndSortModelIDs(entries []upstreamModelDiscoveryEntry) ([]string, error) {
	ids := map[string]struct{}{}
	for _, entry := range entries {
		id := strings.TrimSpace(entry.ID)
		if id == "" {
			id = strings.TrimSpace(entry.Name)
		}
		id = strings.TrimPrefix(id, "models/")
		if id == "" {
			continue
		}
		ids[id] = struct{}{}
	}
	if len(ids) == 0 {
		return nil, errors.New("upstream returned no supported models")
	}

	out := make([]string, 0, len(ids))
	for id := range ids {
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}
