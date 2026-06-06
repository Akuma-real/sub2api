/**
 * Admin Payment API endpoints
 * Handles payment management operations for administrators
 */

import { apiClient } from '../client'
import type {
  DashboardStats,
  PaymentOrder,
  PaymentChannel,
  SubscriptionPlan,
  ProviderInstance,
  VIPLevel,
  VIPUserSummary
} from '@/types/payment'
import type { BasePaginationResponse } from '@/types'

/** Admin-facing payment config returned by GET /admin/payment/config */
export interface AdminPaymentConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  daily_limit: number
  order_timeout_minutes: number
  max_pending_orders: number
  enabled_payment_types: string[]
  balance_disabled: boolean
  balance_recharge_multiplier: number
  load_balance_strategy: string
  product_name_prefix: string
  product_name_suffix: string
  help_image_url: string
  help_text: string
}

/** Fields accepted by PUT /admin/payment/config (all optional via pointer semantics) */
export interface UpdatePaymentConfigRequest {
  enabled?: boolean
  min_amount?: number
  max_amount?: number
  daily_limit?: number
  order_timeout_minutes?: number
  max_pending_orders?: number
  enabled_payment_types?: string[]
  balance_disabled?: boolean
  balance_recharge_multiplier?: number
  load_balance_strategy?: string
  product_name_prefix?: string
  product_name_suffix?: string
  help_image_url?: string
  help_text?: string
}

export const adminPaymentAPI = {
  // ==================== Config ====================

  /** Get payment configuration (admin view) */
  getConfig() {
    return apiClient.get<AdminPaymentConfig>('/admin/payment/config')
  },

  /** Update payment configuration */
  updateConfig(data: UpdatePaymentConfigRequest) {
    return apiClient.put('/admin/payment/config', data)
  },

  // ==================== Dashboard ====================

  /** Get payment dashboard statistics */
  getDashboard(days?: number) {
    return apiClient.get<DashboardStats>('/admin/payment/dashboard', {
      params: days ? { days } : undefined
    })
  },

  // ==================== Orders ====================

  /** Get all orders (paginated, with filters) */
  getOrders(params?: {
    page?: number
    page_size?: number
    status?: string
    payment_type?: string
    user_id?: number
    keyword?: string
    start_date?: string
    end_date?: string
    order_type?: string
  }) {
    return apiClient.get<BasePaginationResponse<PaymentOrder>>('/admin/payment/orders', { params })
  },

  /** Get a specific order by ID */
  getOrder(id: number) {
    return apiClient.get<PaymentOrder>(`/admin/payment/orders/${id}`)
  },

  /** Cancel an order (admin) */
  cancelOrder(id: number) {
    return apiClient.post(`/admin/payment/orders/${id}/cancel`)
  },

  /** Retry recharge for a failed order */
  retryRecharge(id: number) {
    return apiClient.post(`/admin/payment/orders/${id}/retry`)
  },

  /** Process a refund */
  refundOrder(id: number, data: { amount: number; reason: string; deduct_balance?: boolean; force?: boolean }) {
    return apiClient.post(`/admin/payment/orders/${id}/refund`, data)
  },

  // ==================== Channels ====================

  /** Get all payment channels */
  getChannels() {
    return apiClient.get<PaymentChannel[]>('/admin/payment/channels')
  },

  /** Create a payment channel */
  createChannel(data: Partial<PaymentChannel>) {
    return apiClient.post<PaymentChannel>('/admin/payment/channels', data)
  },

  /** Update a payment channel */
  updateChannel(id: number, data: Partial<PaymentChannel>) {
    return apiClient.put<PaymentChannel>(`/admin/payment/channels/${id}`, data)
  },

  /** Delete a payment channel */
  deleteChannel(id: number) {
    return apiClient.delete(`/admin/payment/channels/${id}`)
  },

  // ==================== Subscription Plans ====================

  /** Get all subscription plans */
  getPlans() {
    return apiClient.get<SubscriptionPlan[]>('/admin/payment/plans')
  },

  /** Create a subscription plan */
  createPlan(data: Record<string, unknown>) {
    return apiClient.post<SubscriptionPlan>('/admin/payment/plans', data)
  },

  /** Update a subscription plan */
  updatePlan(id: number, data: Record<string, unknown>) {
    return apiClient.put<SubscriptionPlan>(`/admin/payment/plans/${id}`, data)
  },

  /** Delete a subscription plan */
  deletePlan(id: number) {
    return apiClient.delete(`/admin/payment/plans/${id}`)
  },

  // ==================== VIP Levels ====================

  /** Get all VIP levels */
  getVIPLevels() {
    return apiClient.get<VIPLevel[]>('/admin/payment/vip-levels')
  },

  /** Create a VIP level */
  createVIPLevel(data: Record<string, unknown>) {
    return apiClient.post<VIPLevel>('/admin/payment/vip-levels', data)
  },

  /** Update a VIP level */
  updateVIPLevel(id: number, data: Record<string, unknown>) {
    return apiClient.put<VIPLevel>(`/admin/payment/vip-levels/${id}`, data)
  },

  /** Delete a VIP level */
  deleteVIPLevel(id: number) {
    return apiClient.delete(`/admin/payment/vip-levels/${id}`)
  },

  /** Get user VIP and savings summaries */
  getVIPUsers(params?: { page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<VIPUserSummary>>('/admin/payment/vip-users', { params })
  },

  /** Assign or extend VIP membership for a user */
  assignVIP(data: { user_id: number; vip_level_id: number; days: number; notes?: string }) {
    return apiClient.post('/admin/payment/vip-memberships/assign', data)
  },

  // ==================== Provider Instances ====================

  /** Get all provider instances */
  getProviders() {
    return apiClient.get<ProviderInstance[]>('/admin/payment/providers')
  },

  /** Create a provider instance */
  createProvider(data: Partial<ProviderInstance>) {
    return apiClient.post<ProviderInstance>('/admin/payment/providers', data)
  },

  /** Update a provider instance */
  updateProvider(id: number, data: Partial<ProviderInstance>) {
    return apiClient.put<ProviderInstance>(`/admin/payment/providers/${id}`, data)
  },

  /** Delete a provider instance */
  deleteProvider(id: number) {
    return apiClient.delete(`/admin/payment/providers/${id}`)
  }
}

export default adminPaymentAPI
