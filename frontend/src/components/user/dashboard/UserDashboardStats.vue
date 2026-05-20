<template>
  <!-- Row 1: Core Stats -->
  <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
    <!-- Balance -->
    <div v-if="!isSimple" class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-success/15 p-2">
          <svg class="h-5 w-5 text-success" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z" />
          </svg>
        </div>
        <div>
          <p class="text-xs font-medium text-muted">{{ t('dashboard.balance') }}</p>
          <p class="text-xl font-bold text-success">${{ formatBalance(balance) }}</p>
          <p class="text-xs text-muted">{{ t('common.available') }}</p>
        </div>
      </div>
    </div>

    <!-- API Keys -->
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-surface-card p-2">
          <Icon name="key" size="md" class="text-primary-600" :stroke-width="2" />
        </div>
        <div>
          <p class="text-xs font-medium text-muted">{{ t('dashboard.apiKeys') }}</p>
          <p class="text-xl font-bold text-ink">{{ stats?.total_api_keys || 0 }}</p>
          <p class="text-xs text-success">{{ stats?.active_api_keys || 0 }} {{ t('common.active') }}</p>
        </div>
      </div>
    </div>

    <!-- Today Requests -->
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-success/15 p-2">
          <Icon name="chart" size="md" class="text-success" :stroke-width="2" />
        </div>
        <div>
          <p class="text-xs font-medium text-muted">{{ t('dashboard.todayRequests') }}</p>
          <p class="text-xl font-bold text-ink">{{ stats?.today_requests || 0 }}</p>
          <p class="text-xs text-muted">{{ t('common.total') }}: {{ formatNumber(stats?.total_requests || 0) }}</p>
        </div>
      </div>
    </div>

    <!-- Today Cost -->
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-surface-card p-2">
          <Icon name="dollar" size="md" class="text-accent-amber" :stroke-width="2" />
        </div>
        <div>
          <p class="text-xs font-medium text-muted">{{ t('dashboard.todayCost') }}</p>
          <p class="text-xl font-bold text-ink">
            <span class="text-primary-600" :title="t('dashboard.actual')">${{ formatCost(stats?.today_actual_cost || 0) }}</span>
            <span class="text-sm font-normal text-muted-soft" :title="t('dashboard.standard')"> / ${{ formatCost(stats?.today_cost || 0) }}</span>
          </p>
          <p class="text-xs">
            <span class="text-muted">{{ t('common.total') }}: </span>
            <span class="text-primary-600" :title="t('dashboard.actual')">${{ formatCost(stats?.total_actual_cost || 0) }}</span>
            <span class="text-muted-soft" :title="t('dashboard.standard')"> / ${{ formatCost(stats?.total_cost || 0) }}</span>
          </p>
        </div>
      </div>
    </div>
  </div>

  <!-- Row 2: Token Stats -->
  <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
    <!-- Today Tokens -->
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-accent-amber/15 p-2">
          <Icon name="cube" size="md" class="text-accent-amber" :stroke-width="2" />
        </div>
        <div>
          <p class="text-xs font-medium text-muted">{{ t('dashboard.todayTokens') }}</p>
          <p class="text-xl font-bold text-ink">{{ formatTokens(stats?.today_tokens || 0) }}</p>
          <p class="text-xs text-muted">{{ t('dashboard.input') }}: {{ formatTokens(stats?.today_input_tokens || 0) }} / {{ t('dashboard.output') }}: {{ formatTokens(stats?.today_output_tokens || 0) }}</p>
        </div>
      </div>
    </div>

    <!-- Total Tokens -->
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-surface-card p-2">
          <Icon name="database" size="md" class="text-accent-teal" :stroke-width="2" />
        </div>
        <div>
          <p class="text-xs font-medium text-muted">{{ t('dashboard.totalTokens') }}</p>
          <p class="text-xl font-bold text-ink">{{ formatTokens(stats?.total_tokens || 0) }}</p>
          <p class="text-xs text-muted">{{ t('dashboard.input') }}: {{ formatTokens(stats?.total_input_tokens || 0) }} / {{ t('dashboard.output') }}: {{ formatTokens(stats?.total_output_tokens || 0) }}</p>
        </div>
      </div>
    </div>

    <!-- Performance (RPM/TPM) -->
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-surface-card p-2">
          <Icon name="bolt" size="md" class="text-primary-600" :stroke-width="2" />
        </div>
        <div class="flex-1">
          <p class="text-xs font-medium text-muted">{{ t('dashboard.performance') }}</p>
          <div class="flex items-baseline gap-2">
            <p class="text-xl font-bold text-ink">{{ formatTokens(stats?.rpm || 0) }}</p>
            <span class="text-xs text-muted">RPM</span>
          </div>
          <div class="flex items-baseline gap-2">
            <p class="text-sm font-semibold text-primary-600">{{ formatTokens(stats?.tpm || 0) }}</p>
            <span class="text-xs text-muted">TPM</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Avg Response Time -->
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-surface-card p-2">
          <Icon name="clock" size="md" class="text-error" :stroke-width="2" />
        </div>
        <div>
          <p class="text-xs font-medium text-muted">{{ t('dashboard.avgResponse') }}</p>
          <p class="text-xl font-bold text-ink">{{ formatDuration(stats?.average_duration_ms || 0) }}</p>
          <p class="text-xs text-muted">{{ t('dashboard.averageTime') }}</p>
        </div>
      </div>
    </div>
  </div>

  <!-- Row 3: Per-platform breakdown -->
  <div v-if="!isSimple && platformCards.length > 0" class="card p-4">
    <div class="mb-3 flex items-center justify-between">
      <h3 class="text-sm font-semibold text-ink">{{ t('dashboard.platformBreakdown') }}</h3>
      <span class="text-xs text-muted">
        {{ t('dashboard.platformCount', { count: sortedPlatforms.length }) }}
      </span>
    </div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <div
        v-for="item in platformCards"
        :key="item.platform"
        :class="[
          'rounded-lg border p-3',
          item.isOther
            ? 'border-dashed border-hairline bg-cream'
            : 'border-hairline'
        ]"
      >
        <div class="flex items-center justify-between">
          <span class="text-sm font-semibold text-ink">
            {{ item.isOther ? t('dashboard.platformOther') : platformLabel(item.platform) }}
          </span>
          <span class="font-mono text-sm text-primary-600" :title="t('dashboard.actual')">
            ${{ formatCost(item.total_actual_cost) }}
          </span>
        </div>
        <div class="mt-2 space-y-1 text-xs">
          <div class="flex items-center justify-between">
            <span class="text-muted">{{ t('dashboard.todayCost') }}</span>
            <span class="font-mono text-ink">${{ formatCost(item.today_actual_cost) }}</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-muted">{{ t('dashboard.requests') }}</span>
            <span class="font-mono text-body">
              {{ item.total_requests > 0 ? formatNumber(item.total_requests) : '-' }}
            </span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-muted">{{ t('dashboard.tokens') }}</span>
            <span class="font-mono text-body">
              {{ item.total_tokens > 0 ? formatTokens(item.total_tokens) : '-' }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

const props = defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
}>()
const { t } = useI18n()

const PLATFORM_LABELS: Record<string, string> = {
  anthropic: 'Claude',
  openai: 'OpenAI',
  gemini: 'Gemini',
  antigravity: 'Antigravity'
}

const platformLabel = (p: string) => PLATFORM_LABELS[p] ?? p

const sortedPlatforms = computed(() => {
  const list = props.stats?.by_platform ?? []
  return [...list].sort((a, b) => b.total_actual_cost - a.total_actual_cost)
})

// 处理"各平台之和 < 总值"的差值：后端按平台聚合时过滤了无法归属平台的行
// （group 与 account 都缺 platform）。这里把差值作为"其他"卡片显式展示，
// 避免 Row 1 总值与 Row 3 平台拆分加总对不上、用户困惑。
const OTHER_THRESHOLD = 0.0001
const platformCards = computed(() => {
  const cards: Array<{
    platform: string
    total_actual_cost: number
    today_actual_cost: number
    total_requests: number
    total_tokens: number
    isOther?: boolean
  }> = sortedPlatforms.value.map((p) => ({ ...p }))

  const total = props.stats?.total_actual_cost ?? 0
  const today = props.stats?.today_actual_cost ?? 0
  const sumTotal = cards.reduce((s, c) => s + c.total_actual_cost, 0)
  const sumToday = cards.reduce((s, c) => s + c.today_actual_cost, 0)
  const diffTotal = Math.max(0, total - sumTotal)
  const diffToday = Math.max(0, today - sumToday)

  if (diffTotal > OTHER_THRESHOLD || diffToday > OTHER_THRESHOLD) {
    cards.push({
      platform: '__other__',
      total_actual_cost: diffTotal,
      today_actual_cost: diffToday,
      total_requests: 0,
      total_tokens: 0,
      isOther: true
    })
  }
  return cards
})

const formatBalance = (b: number) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(b)

const formatNumber = (n: number) => n.toLocaleString()
const formatCost = (c: number) => c.toFixed(4)
const formatTokens = (t: number) => {
  if (t >= 1_000_000) return `${(t / 1_000_000).toFixed(1)}M`
  if (t >= 1000) return `${(t / 1000).toFixed(1)}K`
  return t.toString()
}
const formatDuration = (ms: number) => ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`
</script>
