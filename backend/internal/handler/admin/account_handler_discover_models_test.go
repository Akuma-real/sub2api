package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func setupDiscoverModelsRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewAccountHandler(newStubAdminService(), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router.POST("/api/v1/admin/accounts/discover-models", handler.DiscoverUpstreamModels)
	return router
}

func TestAccountHandlerDiscoverUpstreamModels_OpenAICompatible(t *testing.T) {
	var authHeader string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/models", r.URL.Path)
		authHeader = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"data":[{"id":"z-model"},{"id":"a-model"},{"id":"models/gemini-2.5"}]}`))
	}))
	defer upstream.Close()

	body, err := json.Marshal(DiscoverUpstreamModelsRequest{
		Platform: "openai",
		BaseURL:  upstream.URL,
		APIKey:   "sk-test",
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/discover-models", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	setupDiscoverModelsRouter().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "Bearer sk-test", authHeader)

	var resp struct {
		Data DiscoverUpstreamModelsResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, []string{"a-model", "gemini-2.5", "z-model"}, resp.Data.Models)
}

func TestAccountHandlerDiscoverUpstreamModels_GeminiUsesAPIKeyQuery(t *testing.T) {
	var rawQuery string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1beta/models", r.URL.Path)
		rawQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"models":[{"name":"models/gemini-2.5-pro"}]}`))
	}))
	defer upstream.Close()

	body, err := json.Marshal(DiscoverUpstreamModelsRequest{
		Platform: "gemini",
		BaseURL:  upstream.URL,
		APIKey:   "AIza test",
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/discover-models", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	setupDiscoverModelsRouter().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "key=AIza+test", rawQuery)

	var resp struct {
		Data DiscoverUpstreamModelsResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, []string{"gemini-2.5-pro"}, resp.Data.Models)
}

func TestAccountHandlerDiscoverUpstreamModels_UsesConfiguredProxy(t *testing.T) {
	var proxyHit bool
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyHit = true
		require.Equal(t, "upstream.internal", r.URL.Host)
		require.Equal(t, "/v1/models", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":[{"id":"proxied-model"}]}`))
	}))
	defer proxyServer.Close()

	parsedProxyURL, err := url.Parse(proxyServer.URL)
	require.NoError(t, err)
	host, portText, err := net.SplitHostPort(parsedProxyURL.Host)
	require.NoError(t, err)
	port, err := strconv.Atoi(portText)
	require.NoError(t, err)

	adminSvc := newStubAdminService()
	adminSvc.proxies = []service.Proxy{{
		ID:       99,
		Name:     "discovery-proxy",
		Protocol: "http",
		Host:     host,
		Port:     port,
		Status:   service.StatusActive,
	}}

	router := gin.New()
	handler := NewAccountHandler(adminSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router.POST("/api/v1/admin/accounts/discover-models", handler.DiscoverUpstreamModels)

	proxyID := int64(99)
	body, err := json.Marshal(DiscoverUpstreamModelsRequest{
		Platform: "openai",
		BaseURL:  "http://upstream.internal",
		APIKey:   "sk-test",
		ProxyID:  &proxyID,
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/discover-models", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, proxyHit)

	var resp struct {
		Data DiscoverUpstreamModelsResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, []string{"proxied-model"}, resp.Data.Models)
}

func TestRequestUpstreamModelIDs_RedactsQuerySecretsFromTransportErrors(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf(`Get "%s": %w`, r.URL.String(), errors.New("dial failed"))
		}),
	}

	_, err := requestUpstreamModelIDs(
		context.Background(),
		client,
		"https://generativelanguage.googleapis.com/v1beta/models?key=AIza-secret-value",
		nil,
	)

	require.Error(t, err)
	require.NotContains(t, err.Error(), "AIza-secret-value")
	require.Contains(t, err.Error(), "key=[REDACTED]")
}

func TestAccountHandlerDiscoverUpstreamModels_RejectsUnsupportedPlatform(t *testing.T) {
	body, err := json.Marshal(DiscoverUpstreamModelsRequest{
		Platform: "unknown",
		BaseURL:  "https://example.com",
		APIKey:   "sk-test",
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/discover-models", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	setupDiscoverModelsRouter().ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
