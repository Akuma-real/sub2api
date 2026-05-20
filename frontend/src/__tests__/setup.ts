/**
 * Vitest 测试环境设置
 * 提供全局 mock 和测试工具
 */
import { config } from '@vue/test-utils'
import { afterEach, beforeEach, vi } from 'vitest'

function createMemoryStorage(): Storage {
  const values = new Map<string, string>()

  return {
    get length() {
      return values.size
    },
    clear() {
      values.clear()
    },
    getItem(key: string) {
      return values.has(key) ? values.get(key)! : null
    },
    key(index: number) {
      return Array.from(values.keys())[index] ?? null
    },
    removeItem(key: string) {
      values.delete(key)
    },
    setItem(key: string, value: string) {
      values.set(key, String(value))
    }
  }
}

if (typeof globalThis.localStorage === 'undefined' || typeof globalThis.localStorage.getItem !== 'function') {
  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true,
    value: createMemoryStorage()
  })
}

if (typeof window !== 'undefined' && typeof window.localStorage.getItem !== 'function') {
  Object.defineProperty(window, 'localStorage', {
    configurable: true,
    value: globalThis.localStorage
  })
}

// Mock requestIdleCallback (Safari < 15 不支持)
if (typeof globalThis.requestIdleCallback === 'undefined') {
  globalThis.requestIdleCallback = ((callback: IdleRequestCallback) => {
    return window.setTimeout(() => callback({ didTimeout: false, timeRemaining: () => 50 }), 1)
  }) as unknown as typeof requestIdleCallback
}

if (typeof globalThis.cancelIdleCallback === 'undefined') {
  globalThis.cancelIdleCallback = ((id: number) => {
    window.clearTimeout(id)
  }) as unknown as typeof cancelIdleCallback
}

// Mock IntersectionObserver
class MockIntersectionObserver {
  observe = vi.fn()
  disconnect = vi.fn()
  unobserve = vi.fn()
}

globalThis.IntersectionObserver = MockIntersectionObserver as unknown as typeof IntersectionObserver

// Mock ResizeObserver
class MockResizeObserver {
  observe = vi.fn()
  disconnect = vi.fn()
  unobserve = vi.fn()
}

globalThis.ResizeObserver = MockResizeObserver as unknown as typeof ResizeObserver

// Vue Test Utils 全局配置
config.global.stubs = {
  RouterLink: { template: '<a><slot /></a>' },
  'router-link': { template: '<a><slot /></a>' },
}

const originalConsole = {
  error: console.error,
  warn: console.warn,
  debug: console.debug,
}

const expectedConsolePrefixes = [
  'Failed to parse saved user data:',
  'Failed to fetch active subscriptions:',
  'Table load error:',
  '[OpsOpenAITokenStatsCard] Failed to load data',
]

function isExpectedTestConsole(args: unknown[]): boolean {
  const first = args[0]
  return (
    typeof first === 'string' &&
    expectedConsolePrefixes.some((prefix) => first.startsWith(prefix))
  )
}

let errorSpy: ReturnType<typeof vi.spyOn> | null = null
let warnSpy: ReturnType<typeof vi.spyOn> | null = null
let debugSpy: ReturnType<typeof vi.spyOn> | null = null

beforeEach(() => {
  errorSpy = vi.spyOn(console, 'error').mockImplementation((...args: unknown[]) => {
    if (!isExpectedTestConsole(args)) {
      originalConsole.error(...args)
    }
  })
  warnSpy = vi.spyOn(console, 'warn').mockImplementation((...args: unknown[]) => {
    if (!isExpectedTestConsole(args)) {
      originalConsole.warn(...args)
    }
  })
  debugSpy = vi.spyOn(console, 'debug').mockImplementation((...args: unknown[]) => {
    if (!isExpectedTestConsole(args)) {
      originalConsole.debug(...args)
    }
  })
})

afterEach(() => {
  errorSpy?.mockRestore()
  warnSpy?.mockRestore()
  debugSpy?.mockRestore()
  errorSpy = null
  warnSpy = null
  debugSpy = null
})

// 设置全局测试超时
vi.setConfig({ testTimeout: 10000 })
