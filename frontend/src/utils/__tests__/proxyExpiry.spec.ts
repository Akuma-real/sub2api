import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  daysUntil,
  isProxyBackupSelectable,
  isoToLocalDateInput,
  localDateInputEndUnixSeconds,
  proxyExpiryBadgeClass,
  proxyExpiryLabelKey,
  toLocalDateStr,
} from '../proxyExpiry'

// 固定「现在」,按天数构造确定输入:isoInDays(n) 距今正好 n 天
const NOW = new Date('2026-06-02T00:00:00Z')
const isoInDays = (n: number): string => new Date(NOW.getTime() + n * 86400000).toISOString()

beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(NOW)
})
afterEach(() => {
  vi.useRealTimers()
})

describe('daysUntil', () => {
  it('返回距今整天数', () => {
    expect(daysUntil(isoInDays(10))).toBe(10)
    expect(daysUntil(isoInDays(-3))).toBe(-3)
  })
})

describe('local date helpers', () => {
  it('formats a Date using local calendar fields', () => {
    expect(toLocalDateStr(new Date(2026, 5, 9, 12, 30, 0))).toBe('2026-06-09')
  })

  it('converts an ISO timestamp back to the local date input value', () => {
    const local = new Date(2026, 5, 9, 23, 59, 59, 999)
    expect(isoToLocalDateInput(local.toISOString())).toBe('2026-06-09')
  })

  it('converts a date input to local end-of-day unix seconds', () => {
    const expected = Math.floor(new Date(2026, 5, 9, 23, 59, 59, 999).getTime() / 1000)
    expect(localDateInputEndUnixSeconds('2026-06-09')).toBe(expected)
  })
})

describe('isProxyBackupSelectable', () => {
  it('only accepts active proxies that are not expired by time', () => {
    expect(isProxyBackupSelectable({ status: 'active', expires_at: null }, NOW)).toBe(true)
    expect(isProxyBackupSelectable({ status: 'active', expires_at: isoInDays(1) }, NOW)).toBe(true)
    expect(isProxyBackupSelectable({ status: 'inactive', expires_at: isoInDays(1) }, NOW)).toBe(false)
    expect(isProxyBackupSelectable({ status: 'active', expires_at: isoInDays(-1) }, NOW)).toBe(false)
  })
})

describe('proxyExpiryBadgeClass', () => {
  it('status=expired → danger', () => {
    expect(proxyExpiryBadgeClass(isoInDays(30), 'expired')).toBe('badge badge-danger')
  })
  it('≤3 天 → danger（含边界 3）', () => {
    expect(proxyExpiryBadgeClass(isoInDays(2), 'active')).toBe('badge badge-danger')
    expect(proxyExpiryBadgeClass(isoInDays(3), 'active')).toBe('badge badge-danger')
  })
  it('4–7 天 → warning（含边界 7）', () => {
    expect(proxyExpiryBadgeClass(isoInDays(5), 'active')).toBe('badge badge-warning')
    expect(proxyExpiryBadgeClass(isoInDays(7), 'active')).toBe('badge badge-warning')
  })
  it('>7 天 → muted', () => {
    expect(proxyExpiryBadgeClass(isoInDays(30), 'active')).toBe('text-muted')
  })
})

describe('proxyExpiryLabelKey', () => {
  it('status=expired → expired key', () => {
    expect(proxyExpiryLabelKey(isoInDays(30), 'expired')).toEqual({ key: 'admin.proxies.expired' })
  })
  it('已逾期(d<0) → overdueDays', () => {
    expect(proxyExpiryLabelKey(isoInDays(-3), 'active')).toEqual({
      key: 'admin.proxies.overdueDays',
      params: { days: 3 },
    })
  })
  it('≤7 天 → expiringInDays', () => {
    expect(proxyExpiryLabelKey(isoInDays(5), 'active')).toEqual({
      key: 'admin.proxies.expiringInDays',
      params: { days: 5 },
    })
  })
  it('>7 天 → remainingDays', () => {
    expect(proxyExpiryLabelKey(isoInDays(30), 'active')).toEqual({
      key: 'admin.proxies.remainingDays',
      params: { days: 30 },
    })
  })
})
