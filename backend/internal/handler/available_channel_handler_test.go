//go:build unit

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUserAvailableChannel_Unauthenticated401(t *testing.T) {
	// 没有 AuthSubject 注入时，handler 应返回 401 且不触达 service 依赖。
	gin.SetMode(gin.TestMode)
	h := &AvailableChannelHandler{} // nil services — 401 路径不会调用它们
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/channels/available", nil)

	h.List(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFilterUserVisibleGroups_IntersectionOnly(t *testing.T) {
	// 渠道挂在 {g1, g2, g3}，用户只允许 {g1, g3} —— 响应必须仅含 g1/g3。
	groups := []service.AvailableGroupRef{
		{ID: 1, Name: "g1", Platform: "anthropic"},
		{ID: 2, Name: "g2", Platform: "anthropic"},
		{ID: 3, Name: "g3", Platform: "openai"},
	}
	allowed := map[int64]struct{}{1: {}, 3: {}}

	visible := filterUserVisibleGroups(groups, allowed)
	require.Len(t, visible, 2)
	ids := []int64{visible[0].ID, visible[1].ID}
	require.ElementsMatch(t, []int64{1, 3}, ids)
}

func TestToUserSupportedModels_FiltersByAllowedPlatforms(t *testing.T) {
	// 用户可访问分组只覆盖 anthropic；anthropic 平台的模型保留，openai 模型被剔除。
	src := []service.SupportedModel{
		{Name: "claude-sonnet-4-6", Platform: "anthropic", Pricing: nil},
		{Name: "gpt-4o", Platform: "openai", Pricing: nil},
	}
	allowed := map[string]struct{}{"anthropic": {}}
	out := toUserSupportedModels(src, allowed)
	require.Len(t, out, 1)
	require.Equal(t, "claude-sonnet-4-6", out[0].Name)
}

func TestToUserSupportedModels_NilAllowedPlatformsKeepsAll(t *testing.T) {
	// 显式传 nil allowedPlatforms 表示不做过滤。
	src := []service.SupportedModel{
		{Name: "a", Platform: "anthropic"},
		{Name: "b", Platform: "openai"},
	}
	require.Len(t, toUserSupportedModels(src, nil), 2)
}

func TestUserAvailableChannel_FieldWhitelist(t *testing.T) {
	// 通过序列化 userAvailableChannel 结构体验证响应形状：
	// 只有 name / description / platforms；不含管理端字段。
	row := userAvailableChannel{
		Name:        "ch",
		Description: "d",
		Platforms: []userChannelPlatformSection{
			{
				Platform:        "anthropic",
				Groups:          []userAvailableGroup{{ID: 1, Name: "g1", Platform: "anthropic"}},
				SupportedModels: []userSupportedModel{},
			},
		},
	}
	raw, err := json.Marshal(row)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	for _, key := range []string{"id", "status", "billing_model_source", "restrict_models"} {
		_, exists := decoded[key]
		require.Falsef(t, exists, "user DTO must not expose %q", key)
	}
	for _, key := range []string{"name", "description", "platforms"} {
		_, exists := decoded[key]
		require.Truef(t, exists, "user DTO must expose %q", key)
	}

	// 验证 section 的字段（platform / groups / supported_models）。
	rawSection, err := json.Marshal(row.Platforms[0])
	require.NoError(t, err)
	var sectionDecoded map[string]any
	require.NoError(t, json.Unmarshal(rawSection, &sectionDecoded))
	for _, key := range []string{"platform", "groups", "supported_models"} {
		_, exists := sectionDecoded[key]
		require.Truef(t, exists, "platform section must expose %q", key)
	}

	// Group DTO 暴露区分专属/公开、订阅类型、默认倍率所需的字段，
	// 前端据此渲染 GroupBadge 并与 API 密钥页保持一致的视觉。
	rawGroup, err := json.Marshal(row.Platforms[0].Groups[0])
	require.NoError(t, err)
	var groupDecoded map[string]any
	require.NoError(t, json.Unmarshal(rawGroup, &groupDecoded))
	for _, key := range []string{"id", "name", "platform", "subscription_type", "rate_multiplier", "is_exclusive"} {
		_, exists := groupDecoded[key]
		require.Truef(t, exists, "group DTO must expose %q", key)
	}

	// pricing interval 白名单：不应暴露 id / sort_order。
	pricing := toUserPricing(&service.ChannelModelPricing{
		BillingMode: service.BillingModeToken,
		Intervals: []service.PricingInterval{
			{ID: 7, MinTokens: 0, MaxTokens: nil, SortOrder: 3},
		},
	})
	require.NotNil(t, pricing)
	require.Len(t, pricing.Intervals, 1)
	rawIv, err := json.Marshal(pricing.Intervals[0])
	require.NoError(t, err)
	var ivDecoded map[string]any
	require.NoError(t, json.Unmarshal(rawIv, &ivDecoded))
	for _, key := range []string{"id", "pricing_id", "sort_order"} {
		_, exists := ivDecoded[key]
		require.Falsef(t, exists, "user pricing interval must not expose %q", key)
	}
}

func TestBuildPlatformSections_GroupsByPlatform(t *testing.T) {
	// 一个渠道横跨 anthropic / openai / 空平台：应该生成 2 个 section，
	// 按 platform 字母序排序，各自 groups 和 supported_models 只含同平台条目。
	ch := service.AvailableChannel{
		Name: "ch",
		SupportedModels: []service.SupportedModel{
			{Name: "claude-sonnet-4-6", Platform: "anthropic"},
			{Name: "gpt-4o", Platform: "openai"},
		},
	}
	visible := []userAvailableGroup{
		{ID: 1, Name: "g-openai", Platform: "openai"},
		{ID: 2, Name: "g-ant", Platform: "anthropic"},
		{ID: 3, Name: "g-empty", Platform: ""},
	}
	sections := buildPlatformSections(ch, visible)
	require.Len(t, sections, 2)
	require.Equal(t, "anthropic", sections[0].Platform)
	require.Equal(t, "openai", sections[1].Platform)
	require.Len(t, sections[0].Groups, 1)
	require.Equal(t, int64(2), sections[0].Groups[0].ID)
	require.Len(t, sections[0].SupportedModels, 1)
	require.Equal(t, "claude-sonnet-4-6", sections[0].SupportedModels[0].Name)
}

func TestBuildUserModelMarketplace_AggregatesUserVisibleModels(t *testing.T) {
	inputPrice := 0.000002
	outputPrice := 0.00001
	channels := []service.AvailableChannel{
		{
			ID:                 10,
			Name:               "OpenAI Main",
			Description:        "primary",
			Status:             service.StatusActive,
			BillingModelSource: service.BillingModelSourceChannelMapped,
			Groups: []service.AvailableGroupRef{
				{ID: 1, Name: "VIP", Platform: "openai", RateMultiplier: 0.8, IsExclusive: true},
				{ID: 2, Name: "Hidden", Platform: "anthropic", RateMultiplier: 1.2},
			},
			SupportedModels: []service.SupportedModel{
				{
					Name:          "gpt-5.5",
					Platform:      "openai",
					PricingSource: "channel",
					MappedModel:   "gpt-5.5-pro",
					Pricing: &service.ChannelModelPricing{
						BillingMode: service.BillingModeToken,
						InputPrice:  &inputPrice,
						OutputPrice: &outputPrice,
					},
				},
				{Name: "claude-sonnet-4-6", Platform: "anthropic"},
			},
		},
		{
			ID:     11,
			Name:   "Disabled",
			Status: service.StatusDisabled,
			Groups: []service.AvailableGroupRef{
				{ID: 1, Name: "VIP", Platform: "openai", RateMultiplier: 0.8},
			},
			SupportedModels: []service.SupportedModel{{Name: "gpt-hidden", Platform: "openai"}},
		},
	}

	out := buildUserModelMarketplace(
		channels,
		map[int64]struct{}{1: {}},
		map[int64]float64{1: 0.6},
	)

	require.Equal(t, 1, out.Summary.Models)
	require.Equal(t, 1, out.Summary.Platforms)
	require.Equal(t, 1, out.Summary.Channels)
	require.Equal(t, 1, out.Summary.Groups)
	require.Equal(t, 1, out.Summary.PricedModels)
	require.Len(t, out.Models, 1)
	model := out.Models[0]
	require.Equal(t, "gpt-5.5", model.ID)
	require.Equal(t, "openai", model.Platform)
	require.Equal(t, "token", model.BillingMode)
	require.True(t, model.Capabilities.SupportsCachePricing == false)
	require.Len(t, model.Channels, 1)
	require.Equal(t, "gpt-5.5→gpt-5.5-pro", model.Channels[0].Mapping.Chain)
	require.Len(t, model.Channels[0].Groups, 1)
	require.Equal(t, 0.6, model.Channels[0].Groups[0].EffectiveRateMultiplier)
	require.NotNil(t, model.Channels[0].Groups[0].UserRateMultiplier)
	require.Equal(t, 0.6, *model.Channels[0].Groups[0].UserRateMultiplier)
}

func TestFindMarketplaceModel_UsesPlatformAndRawModelID(t *testing.T) {
	models := []userMarketplaceModel{
		{ID: "chat/special", Platform: "openai"},
		{ID: "chat/special", Platform: "anthropic"},
	}

	model, ok := findMarketplaceModel(models, "anthropic", "chat/special")

	require.True(t, ok)
	require.Equal(t, "anthropic", model.Platform)
	require.Equal(t, "chat/special", model.ID)
}
