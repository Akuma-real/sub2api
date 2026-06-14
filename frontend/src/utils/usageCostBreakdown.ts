import type { UsageCostBreakdownAttempt, UsageCostBreakdownSnapshot } from '@/types'

export interface UsageCostBreakdownHost {
  service_tier?: string | null
  vip_discount_multiplier?: number | null
  vip_pre_discount_cost?: number | null
  vip_savings_usd?: number
  dual_protection_enabled?: boolean
  dual_attempt_count?: number
  dual_extra_cost?: number
  actual_cost?: number
  cost_breakdown?: UsageCostBreakdownSnapshot | null
}

const finiteNumber = (value: unknown): number | undefined =>
  typeof value === 'number' && Number.isFinite(value) ? value : undefined

export const usageCostBreakdown = (
  row?: UsageCostBreakdownHost | null
): UsageCostBreakdownSnapshot | null => row?.cost_breakdown ?? null

export const usageFastMode = (row?: UsageCostBreakdownHost | null): string => {
  const mode = String(usageCostBreakdown(row)?.fast?.mode ?? '').trim()
  return mode || 'off'
}

export const usageServiceTier = (row?: UsageCostBreakdownHost | null): string | null => {
  const snapshotTier = usageCostBreakdown(row)?.fast?.service_tier
  const rowTier = row?.service_tier
  const tier = typeof snapshotTier === 'string' ? snapshotTier : rowTier
  return typeof tier === 'string' && tier.trim() ? tier.trim() : null
}

export const usageDualEnabled = (row?: UsageCostBreakdownHost | null): boolean => {
  if (usageCostBreakdown(row)?.dual?.enabled === true) {
    return true
  }
  return Boolean(row?.dual_protection_enabled && usageDualAttemptCount(row) > 0)
}

export const usageDualAttemptCount = (row?: UsageCostBreakdownHost | null): number =>
  finiteNumber(usageCostBreakdown(row)?.dual?.attempt_count) ?? row?.dual_attempt_count ?? 0

export const usageDualExtraCost = (row?: UsageCostBreakdownHost | null): number =>
  finiteNumber(usageCostBreakdown(row)?.dual?.extra_cost) ?? row?.dual_extra_cost ?? 0

export const usageDualPrimaryCost = (row?: UsageCostBreakdownHost | null): number =>
  finiteNumber(usageCostBreakdown(row)?.dual?.primary_cost) ?? 0

export const usageDualSecondaryCost = (row?: UsageCostBreakdownHost | null): number =>
  finiteNumber(usageCostBreakdown(row)?.dual?.secondary_cost) ?? 0

export const usageDualUnsupportedReason = (row?: UsageCostBreakdownHost | null): string => {
  const reason = usageCostBreakdown(row)?.dual?.unsupported_reason
  return typeof reason === 'string' ? reason.trim() : ''
}

export const usageDualBillingBasis = (row?: UsageCostBreakdownHost | null): string => {
  const attempts = usageCostBreakdown(row)?.dual?.attempts
  if (!Array.isArray(attempts)) {
    return ''
  }
  const billedAttempt = attempts.find((attempt: UsageCostBreakdownAttempt) =>
    typeof attempt.billing_basis === 'string' && attempt.billing_basis.trim()
  )
  return billedAttempt?.billing_basis?.trim() ?? ''
}

export const usageDualDisclaimer = (row?: UsageCostBreakdownHost | null): string => {
  const disclaimer = usageCostBreakdown(row)?.dual?.billing_disclaimer
  return typeof disclaimer === 'string' ? disclaimer.trim() : ''
}

export const usageVIPDiscountMultiplier = (row?: UsageCostBreakdownHost | null): number | null => {
  const value = finiteNumber(usageCostBreakdown(row)?.vip?.discount_multiplier) ?? row?.vip_discount_multiplier
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}

export const usageVIPPreDiscountCost = (row?: UsageCostBreakdownHost | null): number | null => {
  const value = finiteNumber(usageCostBreakdown(row)?.vip?.pre_discount_cost) ?? row?.vip_pre_discount_cost
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}

export const usageVIPSavings = (row?: UsageCostBreakdownHost | null): number =>
  finiteNumber(usageCostBreakdown(row)?.vip?.savings_usd) ?? row?.vip_savings_usd ?? 0

export const usageVIPProtectedCost = (row?: UsageCostBreakdownHost | null): number =>
  finiteNumber(usageCostBreakdown(row)?.vip?.protected_cost) ?? 0

export const usageFinalCost = (row?: UsageCostBreakdownHost | null): number =>
  finiteNumber(usageCostBreakdown(row)?.final?.actual_cost) ?? row?.actual_cost ?? 0

export const formatUsageCostUSD = (value: number | null | undefined, digits = 6): string => {
  const numberValue = typeof value === 'number' && Number.isFinite(value) ? value : 0
  return `$${numberValue.toFixed(digits)}`
}
