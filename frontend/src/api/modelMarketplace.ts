import { apiClient } from './client'
import type { BillingMode, BillingModelSource } from '@/constants/channel'
import type { UserSupportedModelPricing } from '@/api/channels'

export interface ModelMarketplaceSummary {
  models: number
  platforms: number
  channels: number
  groups: number
  priced_models: number
  unpriced_models: number
  price_variants: number
}

export interface ModelMarketplaceAvailability {
  status: 'available' | 'partial' | 'unavailable' | 'unknown'
  latest_latency_ms: number | null
  latest_checked_at: string | null
}

export interface ModelMarketplaceCapabilities {
  supports_image: boolean
  supports_cache_pricing: boolean
  has_tiered_pricing: boolean
  has_per_request_pricing: boolean
}

export interface ModelMarketplaceGroup {
  id: number
  name: string
  platform: string
  subscription_type: string
  rate_multiplier: number
  user_rate_multiplier: number | null
  effective_rate_multiplier: number
  is_exclusive: boolean
}

export interface ModelMarketplaceMapping {
  requested_model: string
  mapped_model: string
  billing_model_source: BillingModelSource
  chain: string
}

export interface ModelMarketplaceChannel {
  id: number
  name: string
  description: string
  platform: string
  billing_model_source: BillingModelSource
  pricing: UserSupportedModelPricing | null
  pricing_source: 'channel' | 'global' | 'none' | string
  mapping: ModelMarketplaceMapping
  groups: ModelMarketplaceGroup[]
}

export interface ModelMarketplaceModel {
  id: string
  display_name: string
  platform: string
  billing_mode: BillingMode
  pricing: UserSupportedModelPricing | null
  pricing_source: 'channel' | 'global' | 'none' | string
  availability: ModelMarketplaceAvailability
  capabilities: ModelMarketplaceCapabilities
  channel_count: number
  group_count: number
  price_variant_count: number
  channels: ModelMarketplaceChannel[]
}

export interface ModelMarketplaceResponse {
  summary: ModelMarketplaceSummary
  models: ModelMarketplaceModel[]
}

export async function getMarketplace(options?: {
  signal?: AbortSignal
}): Promise<ModelMarketplaceResponse> {
  const { data } = await apiClient.get<ModelMarketplaceResponse>(
    '/models/marketplace',
    { signal: options?.signal }
  )
  return data
}

export async function getMarketplaceModel(
  model: string,
  platform: string,
  options?: { signal?: AbortSignal }
): Promise<ModelMarketplaceModel> {
  const { data } = await apiClient.get<ModelMarketplaceModel>(
    '/models/marketplace/detail',
    {
      params: { platform, model },
      signal: options?.signal,
    }
  )
  return data
}

export const modelMarketplaceAPI = {
  getMarketplace,
  getMarketplaceModel,
}

export default modelMarketplaceAPI
