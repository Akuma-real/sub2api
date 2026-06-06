/**
 * Payment Store
 * Manages payment configuration, current order state, and subscription plans
 */

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { paymentAPI } from '@/api/payment'
import type { PaymentConfig, PaymentOrder, SubscriptionPlan, CreateOrderRequest, VIPLevel, VIPOverview } from '@/types/payment'

export const usePaymentStore = defineStore('payment', () => {
  // ==================== State ====================

  /** Payment configuration from backend */
  const config = ref<PaymentConfig | null>(null)
  /** Currently active order (for payment flow) */
  const currentOrder = ref<PaymentOrder | null>(null)
  /** Available subscription plans */
  const plans = ref<SubscriptionPlan[]>([])
  /** Available VIP levels */
  const vipLevels = ref<VIPLevel[]>([])
  /** Current user's VIP membership and savings */
  const vipOverview = ref<VIPOverview | null>(null)

  const configLoading = ref(false)
  const configLoaded = ref(false)

  // ==================== Actions ====================

  /** Fetch payment configuration */
  async function fetchConfig(force = false): Promise<PaymentConfig | null> {
    if (configLoaded.value && !force) return config.value
    if (configLoading.value) return config.value

    configLoading.value = true
    try {
      const response = await paymentAPI.getConfig()
      config.value = response.data
      configLoaded.value = true
      return config.value
    } catch (error: unknown) {
      console.error('[payment] Failed to fetch config:', error)
      return null
    } finally {
      configLoading.value = false
    }
  }

  /** Fetch available subscription plans */
  async function fetchPlans(): Promise<SubscriptionPlan[]> {
    try {
      const response = await paymentAPI.getPlans()
      // Backend returns features as newline-separated string; parse to array
      plans.value = (response.data || []).map((p: Omit<SubscriptionPlan, 'features'> & { features: string | string[] }) => ({
        ...p,
        features: typeof p.features === 'string'
          ? p.features.split('\n').map((f: string) => f.trim()).filter(Boolean)
          : (p.features || []),
      }))
      return plans.value
    } catch (error: unknown) {
      console.error('[payment] Failed to fetch plans:', error)
      return []
    }
  }

  /** Fetch available VIP levels */
  async function fetchVIPLevels(): Promise<VIPLevel[]> {
    try {
      const response = await paymentAPI.getVIPLevels()
      vipLevels.value = response.data || []
      return vipLevels.value
    } catch (error: unknown) {
      console.error('[payment] Failed to fetch VIP levels:', error)
      return []
    }
  }

  /** Fetch current user's VIP overview */
  async function fetchVIPOverview(): Promise<VIPOverview | null> {
    try {
      const response = await paymentAPI.getVIPOverview()
      vipOverview.value = response.data
      return vipOverview.value
    } catch (error: unknown) {
      console.error('[payment] Failed to fetch VIP overview:', error)
      return null
    }
  }

  /** Create a new order and set it as current */
  async function createOrder(params: CreateOrderRequest) {
    const response = await paymentAPI.createOrder(params)
    return response.data
  }

  /** Poll order status by ID (read-only, no upstream check) */
  async function pollOrderStatus(orderId: number): Promise<PaymentOrder | null> {
    try {
      const response = await paymentAPI.getOrder(orderId)
      const order = response.data
      if (currentOrder.value?.id === orderId) {
        currentOrder.value = order
      }
      return order
    } catch (error: unknown) {
      console.error('[payment] Failed to poll order status:', error)
      return null
    }
  }

  /** Clear current order state */
  function clearCurrentOrder() {
    currentOrder.value = null
  }

  return {
    config,
    currentOrder,
    plans,
    vipLevels,
    vipOverview,
    configLoading,
    configLoaded,
    fetchConfig,
    fetchPlans,
    fetchVIPLevels,
    fetchVIPOverview,
    createOrder,
    pollOrderStatus,
    clearCurrentOrder
  }
})
