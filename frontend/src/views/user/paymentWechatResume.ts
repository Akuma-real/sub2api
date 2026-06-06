import type { LocationQuery, LocationQueryRaw } from 'vue-router'
import type { OrderType, SubscriptionPlan, VIPLevel } from '@/types/payment'
import { normalizeVisibleMethod } from '@/components/payment/paymentFlow'

export interface ParsedWechatResumeRoute {
  orderAmount: number
  orderType: OrderType
  paymentType: string
  planId?: number
  vipLevelId?: number
  openid?: string
  wechatResumeToken?: string
}

function readQueryString(query: LocationQuery, key: string): string {
  const value = query[key]
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

export function hasWechatResumeQuery(query: LocationQuery): boolean {
  if (readQueryString(query, 'wechat_resume') === '1') {
    return true
  }
  return readQueryString(query, 'wechat_resume_token') !== ''
    || readQueryString(query, 'openid') !== ''
}

export function parseWechatResumeRoute(
  query: LocationQuery,
  plans: SubscriptionPlan[],
  vipLevels: VIPLevel[] = [],
  fallbackBalanceAmount: number,
): ParsedWechatResumeRoute | null {
  if (!hasWechatResumeQuery(query)) {
    return null
  }

  const wechatResumeToken = readQueryString(query, 'wechat_resume_token')
  const paymentType = normalizeVisibleMethod(readQueryString(query, 'payment_type')) || 'wxpay'
  const planId = Number.parseInt(readQueryString(query, 'plan_id'), 10)
  const hasPlanId = Number.isFinite(planId) && planId > 0
  const vipLevelId = Number.parseInt(readQueryString(query, 'vip_level_id'), 10)
  const hasVIPLevelId = Number.isFinite(vipLevelId) && vipLevelId > 0
  const rawOrderType = readQueryString(query, 'order_type')
  const orderType: OrderType = rawOrderType === 'vip' || hasVIPLevelId
    ? 'vip'
    : (rawOrderType === 'subscription' || hasPlanId ? 'subscription' : 'balance')

  if (wechatResumeToken) {
    return {
      wechatResumeToken,
      paymentType,
      orderType,
      orderAmount: 0,
      planId: hasPlanId ? planId : undefined,
      vipLevelId: hasVIPLevelId ? vipLevelId : undefined,
    }
  }

  const openid = readQueryString(query, 'openid')
  if (!openid) {
    return null
  }

  const rawAmount = Number.parseFloat(readQueryString(query, 'amount'))
  const orderAmount = Number.isFinite(rawAmount) && rawAmount > 0
    ? rawAmount
    : (orderType === 'subscription'
      ? (plans.find(plan => plan.id === planId)?.price ?? 0)
      : orderType === 'vip'
        ? (vipLevels.find(level => level.id === vipLevelId)?.price ?? 0)
      : fallbackBalanceAmount)

  return {
    openid,
    paymentType,
    orderType,
    orderAmount,
    planId: hasPlanId ? planId : undefined,
    vipLevelId: hasVIPLevelId ? vipLevelId : undefined,
  }
}

export function stripWechatResumeQuery(query: LocationQuery): LocationQueryRaw {
  const nextQuery: LocationQueryRaw = { ...query }
  delete nextQuery.wechat_resume
  delete nextQuery.wechat_resume_token
  delete nextQuery.openid
  delete nextQuery.state
  delete nextQuery.scope
  delete nextQuery.payment_type
  delete nextQuery.amount
  delete nextQuery.order_type
  delete nextQuery.plan_id
  delete nextQuery.vip_level_id
  return nextQuery
}
