package handler

import (
	"sort"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AvailableChannelHandler 处理用户侧「可用渠道」查询。
//
// 用户侧接口委托 ChannelService.ListAvailable，并在返回前做三层过滤：
//  1. 行过滤：只保留状态为 Active 且与当前用户可访问分组有交集的渠道；
//  2. 分组过滤：渠道的 Groups 只保留用户可访问的那些；
//  3. 平台过滤：渠道的 SupportedModels 只保留平台在用户可见 Groups 中出现过的模型，
//     防止"渠道同时挂在 antigravity / anthropic 两个平台的分组上，用户只访问
//     antigravity，却看到 anthropic 模型"这类跨平台信息泄漏；
//  4. 字段白名单：仅返回用户需要的字段（省略 BillingModelSource / RestrictModels
//     / 内部 ID / Status 等管理字段）。
type AvailableChannelHandler struct {
	channelService *service.ChannelService
	apiKeyService  *service.APIKeyService
	settingService *service.SettingService
}

// NewAvailableChannelHandler 创建用户侧可用渠道 handler。
func NewAvailableChannelHandler(
	channelService *service.ChannelService,
	apiKeyService *service.APIKeyService,
	settingService *service.SettingService,
) *AvailableChannelHandler {
	return &AvailableChannelHandler{
		channelService: channelService,
		apiKeyService:  apiKeyService,
		settingService: settingService,
	}
}

// featureEnabled 返回 available-channels 开关是否启用。默认关闭（opt-in）。
func (h *AvailableChannelHandler) featureEnabled(c *gin.Context) bool {
	if h.settingService == nil {
		return false
	}
	return h.settingService.GetAvailableChannelsRuntime(c.Request.Context()).Enabled
}

// userAvailableGroup 用户可见的分组概要（白名单字段）。
//
// 前端据此区分专属 vs 公开分组（IsExclusive）、订阅 vs 标准分组（SubscriptionType，
// 订阅视觉加深），并用 RateMultiplier 作为默认倍率；用户专属倍率前端走
// /groups/rates，和 API 密钥页面保持一致。
type userAvailableGroup struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Platform         string  `json:"platform"`
	SubscriptionType string  `json:"subscription_type"`
	RateMultiplier   float64 `json:"rate_multiplier"`
	IsExclusive      bool    `json:"is_exclusive"`
}

// userSupportedModelPricing 用户可见的定价字段白名单。
type userSupportedModelPricing struct {
	BillingMode      string                   `json:"billing_mode"`
	InputPrice       *float64                 `json:"input_price"`
	OutputPrice      *float64                 `json:"output_price"`
	CacheWritePrice  *float64                 `json:"cache_write_price"`
	CacheReadPrice   *float64                 `json:"cache_read_price"`
	ImageOutputPrice *float64                 `json:"image_output_price"`
	PerRequestPrice  *float64                 `json:"per_request_price"`
	Intervals        []userPricingIntervalDTO `json:"intervals"`
}

// userPricingIntervalDTO 定价区间白名单（去掉内部 ID、SortOrder 等前端不渲染的字段）。
type userPricingIntervalDTO struct {
	MinTokens       int      `json:"min_tokens"`
	MaxTokens       *int     `json:"max_tokens"`
	TierLabel       string   `json:"tier_label,omitempty"`
	InputPrice      *float64 `json:"input_price"`
	OutputPrice     *float64 `json:"output_price"`
	CacheWritePrice *float64 `json:"cache_write_price"`
	CacheReadPrice  *float64 `json:"cache_read_price"`
	PerRequestPrice *float64 `json:"per_request_price"`
}

// userSupportedModel 用户可见的支持模型条目。
type userSupportedModel struct {
	Name     string                     `json:"name"`
	Platform string                     `json:"platform"`
	Pricing  *userSupportedModelPricing `json:"pricing"`
}

type userModelMarketplaceSummary struct {
	Models         int `json:"models"`
	Platforms      int `json:"platforms"`
	Channels       int `json:"channels"`
	Groups         int `json:"groups"`
	PricedModels   int `json:"priced_models"`
	UnpricedModels int `json:"unpriced_models"`
	PriceVariants  int `json:"price_variants"`
}

type userModelMarketplaceResponse struct {
	Summary userModelMarketplaceSummary `json:"summary"`
	Models  []userMarketplaceModel      `json:"models"`
}

type userMarketplaceModel struct {
	ID                string                        `json:"id"`
	DisplayName       string                        `json:"display_name"`
	Platform          string                        `json:"platform"`
	BillingMode       string                        `json:"billing_mode"`
	Pricing           *userSupportedModelPricing    `json:"pricing"`
	PricingSource     string                        `json:"pricing_source"`
	Availability      userModelAvailability         `json:"availability"`
	Capabilities      userModelCapabilities         `json:"capabilities"`
	ChannelCount      int                           `json:"channel_count"`
	GroupCount        int                           `json:"group_count"`
	PriceVariantCount int                           `json:"price_variant_count"`
	Channels          []userMarketplaceModelChannel `json:"channels"`
}

type userModelAvailability struct {
	Status          string  `json:"status"`
	LatestLatencyMS *int    `json:"latest_latency_ms"`
	LatestCheckedAt *string `json:"latest_checked_at"`
}

type userModelCapabilities struct {
	SupportsImage        bool `json:"supports_image"`
	SupportsCachePricing bool `json:"supports_cache_pricing"`
	HasTieredPricing     bool `json:"has_tiered_pricing"`
	HasPerRequestPricing bool `json:"has_per_request_pricing"`
}

type userMarketplaceModelChannel struct {
	ID                 int64                       `json:"id"`
	Name               string                      `json:"name"`
	Description        string                      `json:"description"`
	Platform           string                      `json:"platform"`
	BillingModelSource string                      `json:"billing_model_source"`
	Pricing            *userSupportedModelPricing  `json:"pricing"`
	PricingSource      string                      `json:"pricing_source"`
	Mapping            userMarketplaceModelMapping `json:"mapping"`
	Groups             []userMarketplaceGroup      `json:"groups"`
}

type userMarketplaceModelMapping struct {
	RequestedModel     string `json:"requested_model"`
	MappedModel        string `json:"mapped_model"`
	BillingModelSource string `json:"billing_model_source"`
	Chain              string `json:"chain"`
}

type userMarketplaceGroup struct {
	ID                      int64    `json:"id"`
	Name                    string   `json:"name"`
	Platform                string   `json:"platform"`
	SubscriptionType        string   `json:"subscription_type"`
	RateMultiplier          float64  `json:"rate_multiplier"`
	UserRateMultiplier      *float64 `json:"user_rate_multiplier"`
	EffectiveRateMultiplier float64  `json:"effective_rate_multiplier"`
	IsExclusive             bool     `json:"is_exclusive"`
}

// userChannelPlatformSection 单渠道内某个平台的子视图：用户可见的分组 + 该平台
// 支持的模型。按 platform 聚合后让前端可以把渠道名作为 row-group 一次渲染，
// 后面的平台行按 sections 顺序铺开。
type userChannelPlatformSection struct {
	Platform        string               `json:"platform"`
	Groups          []userAvailableGroup `json:"groups"`
	SupportedModels []userSupportedModel `json:"supported_models"`
}

// userAvailableChannel 用户可见的渠道条目（白名单字段）。
//
// 每个渠道聚合为一条记录，内嵌 platforms 子数组：每个 section 对应一个平台，
// 包含该平台的 groups 和 supported_models。
type userAvailableChannel struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Platforms   []userChannelPlatformSection `json:"platforms"`
}

// List 列出当前用户可见的「可用渠道」。
// GET /api/v1/channels/available
func (h *AvailableChannelHandler) List(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Feature 未启用时返回空数组（不暴露渠道信息）。检查放在认证之后，
	// 保持与未开关前的 401 行为一致：未登录先 401，登录后再按开关决定。
	if !h.featureEnabled(c) {
		response.Success(c, []userAvailableChannel{})
		return
	}

	userGroups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	allowedGroupIDs := make(map[int64]struct{}, len(userGroups))
	for i := range userGroups {
		allowedGroupIDs[userGroups[i].ID] = struct{}{}
	}

	channels, err := h.channelService.ListAvailable(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]userAvailableChannel, 0, len(channels))
	for _, ch := range channels {
		if ch.Status != service.StatusActive {
			continue
		}
		visibleGroups := filterUserVisibleGroups(ch.Groups, allowedGroupIDs)
		if len(visibleGroups) == 0 {
			continue
		}
		sections := buildPlatformSections(ch, visibleGroups)
		if len(sections) == 0 {
			continue
		}
		out = append(out, userAvailableChannel{
			Name:        ch.Name,
			Description: ch.Description,
			Platforms:   sections,
		})
	}

	response.Success(c, out)
}

// ModelMarketplace 列出当前用户可见的模型广场。
// GET /api/v1/models/marketplace
func (h *AvailableChannelHandler) ModelMarketplace(c *gin.Context) {
	resp, ok := h.buildModelMarketplace(c)
	if !ok {
		return
	}
	response.Success(c, resp)
}

// ModelMarketplaceDetail 返回某个模型在当前用户权限下的完整可用渠道与价格视图。
// GET /api/v1/models/marketplace/detail?platform=openai&model=gpt-5.5
func (h *AvailableChannelHandler) ModelMarketplaceDetail(c *gin.Context) {
	platform := strings.TrimSpace(c.Query("platform"))
	modelID := strings.TrimSpace(c.Query("model"))
	if platform == "" || modelID == "" {
		response.BadRequest(c, "platform and model are required")
		return
	}
	resp, ok := h.buildModelMarketplace(c)
	if !ok {
		return
	}
	if model, ok := findMarketplaceModel(resp.Models, platform, modelID); ok {
		response.Success(c, model)
		return
	}
	response.NotFound(c, "Model not found")
}

func findMarketplaceModel(
	models []userMarketplaceModel,
	platform string,
	modelID string,
) (userMarketplaceModel, bool) {
	for _, model := range models {
		if strings.EqualFold(model.Platform, platform) && strings.EqualFold(model.ID, modelID) {
			return model, true
		}
	}
	return userMarketplaceModel{}, false
}

func (h *AvailableChannelHandler) buildModelMarketplace(c *gin.Context) (userModelMarketplaceResponse, bool) {
	empty := userModelMarketplaceResponse{Models: []userMarketplaceModel{}}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return empty, false
	}
	if !h.featureEnabled(c) {
		response.Success(c, empty)
		return empty, false
	}

	userGroups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return empty, false
	}
	allowedGroupIDs := make(map[int64]struct{}, len(userGroups))
	for i := range userGroups {
		allowedGroupIDs[userGroups[i].ID] = struct{}{}
	}
	userRates, err := h.apiKeyService.GetUserGroupRates(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return empty, false
	}
	if userRates == nil {
		userRates = map[int64]float64{}
	}

	channels, err := h.channelService.ListAvailable(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return empty, false
	}
	resp := buildUserModelMarketplace(channels, allowedGroupIDs, userRates)
	return resp, true
}

func buildUserModelMarketplace(
	channels []service.AvailableChannel,
	allowedGroupIDs map[int64]struct{},
	userRates map[int64]float64,
) userModelMarketplaceResponse {
	type modelKey struct {
		platform string
		name     string
	}
	models := make(map[modelKey]*userMarketplaceModel)
	platforms := make(map[string]struct{})
	channelsSeen := make(map[int64]struct{})
	groupsSeen := make(map[int64]struct{})
	priceVariants := 0

	for _, ch := range channels {
		if ch.Status != service.StatusActive {
			continue
		}
		visibleGroups := filterUserVisibleGroups(ch.Groups, allowedGroupIDs)
		if len(visibleGroups) == 0 {
			continue
		}
		sections := buildPlatformSections(ch, visibleGroups)
		for _, section := range sections {
			groups := marketplaceGroups(section.Groups, userRates)
			if len(groups) == 0 {
				continue
			}
			platforms[section.Platform] = struct{}{}
			channelsSeen[ch.ID] = struct{}{}
			for _, g := range groups {
				groupsSeen[g.ID] = struct{}{}
			}
			for _, model := range ch.SupportedModels {
				if model.Platform != section.Platform {
					continue
				}
				key := modelKey{platform: model.Platform, name: strings.ToLower(model.Name)}
				row, exists := models[key]
				if !exists {
					row = &userMarketplaceModel{
						ID:            model.Name,
						DisplayName:   model.Name,
						Platform:      model.Platform,
						BillingMode:   marketplaceBillingMode(model.Pricing),
						Pricing:       toUserPricing(model.Pricing),
						PricingSource: normalizePricingSource(model.PricingSource, model.Pricing),
						Availability:  userModelAvailability{Status: "unknown"},
						Capabilities:  marketplaceCapabilities(model.Pricing),
						Channels:      []userMarketplaceModelChannel{},
					}
					models[key] = row
				} else {
					row.Capabilities = mergeCapabilities(row.Capabilities, marketplaceCapabilities(model.Pricing))
					if row.Pricing == nil && model.Pricing != nil {
						row.Pricing = toUserPricing(model.Pricing)
						row.PricingSource = normalizePricingSource(model.PricingSource, model.Pricing)
						row.BillingMode = marketplaceBillingMode(model.Pricing)
					}
				}
				row.Channels = append(row.Channels, userMarketplaceModelChannel{
					ID:                 ch.ID,
					Name:               ch.Name,
					Description:        ch.Description,
					Platform:           section.Platform,
					BillingModelSource: ch.BillingModelSource,
					Pricing:            toUserPricing(model.Pricing),
					PricingSource:      normalizePricingSource(model.PricingSource, model.Pricing),
					Mapping:            marketplaceMapping(model, ch.BillingModelSource),
					Groups:             groups,
				})
				if model.Pricing != nil {
					priceVariants++
				}
			}
		}
	}

	out := make([]userMarketplaceModel, 0, len(models))
	priced := 0
	for _, row := range models {
		groupIDs := make(map[int64]struct{})
		for _, ch := range row.Channels {
			for _, g := range ch.Groups {
				groupIDs[g.ID] = struct{}{}
			}
		}
		row.ChannelCount = len(row.Channels)
		row.GroupCount = len(groupIDs)
		row.PriceVariantCount = countPriceVariants(row.Channels)
		if row.Pricing != nil {
			priced++
		}
		sort.SliceStable(row.Channels, func(i, j int) bool {
			if row.Channels[i].Name != row.Channels[j].Name {
				return strings.ToLower(row.Channels[i].Name) < strings.ToLower(row.Channels[j].Name)
			}
			return row.Channels[i].Platform < row.Channels[j].Platform
		})
		out = append(out, *row)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Platform != out[j].Platform {
			return out[i].Platform < out[j].Platform
		}
		return strings.ToLower(out[i].ID) < strings.ToLower(out[j].ID)
	})
	return userModelMarketplaceResponse{
		Summary: userModelMarketplaceSummary{
			Models:         len(out),
			Platforms:      len(platforms),
			Channels:       len(channelsSeen),
			Groups:         len(groupsSeen),
			PricedModels:   priced,
			UnpricedModels: len(out) - priced,
			PriceVariants:  priceVariants,
		},
		Models: out,
	}
}

// buildPlatformSections 把一个渠道按 visibleGroups 的平台集合拆成有序的 section 列表：
// 每个 section 对应一个平台，只包含该平台的 groups 和 supported_models。
// 输出按 platform 字母序稳定排序，便于前端等效比较与回归测试。
func buildPlatformSections(
	ch service.AvailableChannel,
	visibleGroups []userAvailableGroup,
) []userChannelPlatformSection {
	groupsByPlatform := make(map[string][]userAvailableGroup, 4)
	for _, g := range visibleGroups {
		if g.Platform == "" {
			continue
		}
		groupsByPlatform[g.Platform] = append(groupsByPlatform[g.Platform], g)
	}
	if len(groupsByPlatform) == 0 {
		return nil
	}

	platforms := make([]string, 0, len(groupsByPlatform))
	for p := range groupsByPlatform {
		platforms = append(platforms, p)
	}
	sort.Strings(platforms)

	sections := make([]userChannelPlatformSection, 0, len(platforms))
	for _, platform := range platforms {
		platformSet := map[string]struct{}{platform: {}}
		sections = append(sections, userChannelPlatformSection{
			Platform:        platform,
			Groups:          groupsByPlatform[platform],
			SupportedModels: toUserSupportedModels(ch.SupportedModels, platformSet),
		})
	}
	return sections
}

// filterUserVisibleGroups 仅保留用户可访问的分组。
func filterUserVisibleGroups(
	groups []service.AvailableGroupRef,
	allowed map[int64]struct{},
) []userAvailableGroup {
	visible := make([]userAvailableGroup, 0, len(groups))
	for _, g := range groups {
		if _, ok := allowed[g.ID]; !ok {
			continue
		}
		visible = append(visible, userAvailableGroup{
			ID:               g.ID,
			Name:             g.Name,
			Platform:         g.Platform,
			SubscriptionType: g.SubscriptionType,
			RateMultiplier:   g.RateMultiplier,
			IsExclusive:      g.IsExclusive,
		})
	}
	return visible
}

// toUserSupportedModels 将 service 层支持模型转换为用户 DTO（字段白名单）。
// 仅保留平台在 allowedPlatforms 中的条目，防止跨平台模型信息泄漏。
// allowedPlatforms 为 nil 时不做平台过滤（保留全部，供测试或明确无过滤场景使用）。
func toUserSupportedModels(
	src []service.SupportedModel,
	allowedPlatforms map[string]struct{},
) []userSupportedModel {
	out := make([]userSupportedModel, 0, len(src))
	for i := range src {
		m := src[i]
		if allowedPlatforms != nil {
			if _, ok := allowedPlatforms[m.Platform]; !ok {
				continue
			}
		}
		out = append(out, userSupportedModel{
			Name:     m.Name,
			Platform: m.Platform,
			Pricing:  toUserPricing(m.Pricing),
		})
	}
	return out
}

// toUserPricing 将 service 层定价转换为用户 DTO；入参为 nil 时返回 nil。
func toUserPricing(p *service.ChannelModelPricing) *userSupportedModelPricing {
	if p == nil {
		return nil
	}
	intervals := make([]userPricingIntervalDTO, 0, len(p.Intervals))
	for _, iv := range p.Intervals {
		intervals = append(intervals, userPricingIntervalDTO{
			MinTokens:       iv.MinTokens,
			MaxTokens:       iv.MaxTokens,
			TierLabel:       iv.TierLabel,
			InputPrice:      iv.InputPrice,
			OutputPrice:     iv.OutputPrice,
			CacheWritePrice: iv.CacheWritePrice,
			CacheReadPrice:  iv.CacheReadPrice,
			PerRequestPrice: iv.PerRequestPrice,
		})
	}
	billingMode := string(p.BillingMode)
	if billingMode == "" {
		billingMode = string(service.BillingModeToken)
	}
	return &userSupportedModelPricing{
		BillingMode:      billingMode,
		InputPrice:       p.InputPrice,
		OutputPrice:      p.OutputPrice,
		CacheWritePrice:  p.CacheWritePrice,
		CacheReadPrice:   p.CacheReadPrice,
		ImageOutputPrice: p.ImageOutputPrice,
		PerRequestPrice:  p.PerRequestPrice,
		Intervals:        intervals,
	}
}

func marketplaceGroups(groups []userAvailableGroup, userRates map[int64]float64) []userMarketplaceGroup {
	out := make([]userMarketplaceGroup, 0, len(groups))
	for _, g := range groups {
		effective := g.RateMultiplier
		var userRate *float64
		if rate, ok := userRates[g.ID]; ok {
			r := rate
			userRate = &r
			effective = r
		}
		out = append(out, userMarketplaceGroup{
			ID:                      g.ID,
			Name:                    g.Name,
			Platform:                g.Platform,
			SubscriptionType:        g.SubscriptionType,
			RateMultiplier:          g.RateMultiplier,
			UserRateMultiplier:      userRate,
			EffectiveRateMultiplier: effective,
			IsExclusive:             g.IsExclusive,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].IsExclusive != out[j].IsExclusive {
			return out[i].IsExclusive
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

func marketplaceBillingMode(p *service.ChannelModelPricing) string {
	if p == nil || p.BillingMode == "" {
		return string(service.BillingModeToken)
	}
	return string(p.BillingMode)
}

func normalizePricingSource(source string, p *service.ChannelModelPricing) string {
	if p == nil {
		return "none"
	}
	if strings.TrimSpace(source) == "" {
		return "channel"
	}
	return source
}

func marketplaceCapabilities(p *service.ChannelModelPricing) userModelCapabilities {
	if p == nil {
		return userModelCapabilities{}
	}
	return userModelCapabilities{
		SupportsImage:        p.ImageOutputPrice != nil || p.BillingMode == service.BillingModeImage,
		SupportsCachePricing: p.CacheWritePrice != nil || p.CacheReadPrice != nil,
		HasTieredPricing:     len(p.Intervals) > 0,
		HasPerRequestPricing: p.PerRequestPrice != nil || p.BillingMode == service.BillingModePerRequest,
	}
}

func mergeCapabilities(a, b userModelCapabilities) userModelCapabilities {
	return userModelCapabilities{
		SupportsImage:        a.SupportsImage || b.SupportsImage,
		SupportsCachePricing: a.SupportsCachePricing || b.SupportsCachePricing,
		HasTieredPricing:     a.HasTieredPricing || b.HasTieredPricing,
		HasPerRequestPricing: a.HasPerRequestPricing || b.HasPerRequestPricing,
	}
}

func marketplaceMapping(model service.SupportedModel, billingModelSource string) userMarketplaceModelMapping {
	mapped := model.MappedModel
	chain := model.Name
	if mapped != "" && !strings.EqualFold(mapped, model.Name) {
		chain = model.Name + "→" + mapped
	}
	return userMarketplaceModelMapping{
		RequestedModel:     model.Name,
		MappedModel:        mapped,
		BillingModelSource: billingModelSource,
		Chain:              chain,
	}
}

func countPriceVariants(channels []userMarketplaceModelChannel) int {
	seen := make(map[string]struct{})
	for _, ch := range channels {
		if ch.Pricing == nil {
			continue
		}
		key := ch.PricingSource + "|" + ch.Pricing.BillingMode
		if ch.Pricing.InputPrice != nil {
			key += "|i:" + floatKey(*ch.Pricing.InputPrice)
		}
		if ch.Pricing.OutputPrice != nil {
			key += "|o:" + floatKey(*ch.Pricing.OutputPrice)
		}
		if ch.Pricing.CacheWritePrice != nil {
			key += "|cw:" + floatKey(*ch.Pricing.CacheWritePrice)
		}
		if ch.Pricing.CacheReadPrice != nil {
			key += "|cr:" + floatKey(*ch.Pricing.CacheReadPrice)
		}
		if ch.Pricing.ImageOutputPrice != nil {
			key += "|img:" + floatKey(*ch.Pricing.ImageOutputPrice)
		}
		if ch.Pricing.PerRequestPrice != nil {
			key += "|req:" + floatKey(*ch.Pricing.PerRequestPrice)
		}
		seen[key] = struct{}{}
	}
	return len(seen)
}

func floatKey(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}
