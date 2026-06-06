/**
 * Redeem code API endpoints
 * Handles redeem code redemption for users
 */

import { apiClient } from './client'
import type { RedeemCode, RedeemCodeRequest } from '@/types'
import type { VIPLevel } from '@/types/payment'

export interface RedeemHistoryItem {
  id: number
  code: string
  type: string
  value: number
  status: string
  used_at: string
  created_at: string
  // Notes from admin for admin_balance/admin_concurrency types
  notes?: string
  // Subscription-specific fields
  group_id?: number
  validity_days?: number
  group?: {
    id: number
    name: string
  }
  vip_level_id?: number
  vip_days?: number
  vip_level?: VIPLevel
}

/**
 * Redeem a code
 * @param code - Redeem code string
 * @returns Redeemed code snapshot
 */
export async function redeem(code: string): Promise<RedeemCode> {
  const payload: RedeemCodeRequest = { code }

  const { data } = await apiClient.post<RedeemCode>('/redeem', payload)

  return data
}

/**
 * Get user's redemption history
 * @returns List of redeemed codes
 */
export async function getHistory(): Promise<RedeemHistoryItem[]> {
  const { data } = await apiClient.get<RedeemHistoryItem[]>('/redeem/history')
  return data
}

export const redeemAPI = {
  redeem,
  getHistory
}

export default redeemAPI
