<template>
  <div
    :class="[
      'group relative flex h-full flex-col overflow-hidden rounded-lg border border-hairline bg-canvas p-8 shadow-card transition-all',
      'hover:-translate-y-0.5 hover:shadow-card-hover',
    ]"
  >
    <div class="flex flex-col gap-4 sm:min-h-[8.25rem] sm:flex-row sm:items-start sm:justify-between sm:gap-6">
      <div class="min-w-0 flex-1">
        <div class="flex flex-wrap items-center gap-2">
          <h3 class="min-w-0 break-words font-display text-lg font-medium leading-tight text-ink">
            {{ plan.name }}
          </h3>
          <span class="shrink-0 rounded-full border border-hairline bg-surface-card px-2.5 py-0.5 text-[11px] font-medium text-muted">
            {{ pLabel }}
          </span>
        </div>
        <p v-if="plan.description" class="mt-3 line-clamp-2 text-sm leading-relaxed text-muted">
          {{ plan.description }}
        </p>
        <div v-else class="mt-3 hidden min-h-[2.75rem] sm:block" aria-hidden="true"></div>
      </div>

      <div class="w-full shrink-0 text-left sm:w-auto sm:text-right">
        <div class="flex flex-wrap items-baseline gap-x-2 gap-y-1 sm:justify-end">
          <span v-if="originalPriceDisplay" class="whitespace-nowrap text-sm text-muted-soft line-through">
            {{ originalPriceDisplay }}
          </span>
          <span class="whitespace-nowrap font-display text-3xl font-medium leading-none text-ink sm:text-4xl">
            {{ priceDisplay }}
          </span>
          <span class="whitespace-nowrap text-sm text-muted">
            / {{ validitySuffix }}
          </span>
        </div>
        <p v-if="discountText" class="mt-1 text-xs font-medium text-primary-700 sm:text-right">
          {{ discountText }}
        </p>
      </div>
    </div>

    <!-- Quota + scope summary -->
    <div class="mt-6 rounded-lg border border-hairline bg-surface-card p-4 text-sm sm:min-h-[4.75rem]">
      <div class="grid gap-3 sm:grid-cols-2">
        <div class="flex items-center justify-between">
          <span class="text-muted-soft">{{ t('payment.planCard.rate') }}</span>
          <span class="font-medium text-body-strong">{{ rateDisplay }}</span>
        </div>
        <div v-if="plan.daily_limit_usd != null" class="flex items-center justify-between">
          <span class="text-muted-soft">{{ t('payment.planCard.dailyLimit') }}</span>
          <span class="font-medium text-body-strong">${{ plan.daily_limit_usd }}</span>
        </div>
        <div v-if="plan.weekly_limit_usd != null" class="flex items-center justify-between">
          <span class="text-muted-soft">{{ t('payment.planCard.weeklyLimit') }}</span>
          <span class="font-medium text-body-strong">${{ plan.weekly_limit_usd }}</span>
        </div>
        <div v-if="plan.monthly_limit_usd != null" class="flex items-center justify-between">
          <span class="text-muted-soft">{{ t('payment.planCard.monthlyLimit') }}</span>
          <span class="font-medium text-body-strong">${{ plan.monthly_limit_usd }}</span>
        </div>
        <div v-if="plan.daily_limit_usd == null && plan.weekly_limit_usd == null && plan.monthly_limit_usd == null" class="flex items-center justify-between">
          <span class="text-muted-soft">{{ t('payment.planCard.quota') }}</span>
          <span class="font-medium text-body-strong">{{ t('payment.planCard.unlimited') }}</span>
        </div>
        <div v-if="modelScopeLabels.length > 0" class="col-span-2 flex items-center justify-between">
          <span class="text-muted-soft">{{ t('payment.planCard.models') }}</span>
          <div class="flex flex-wrap justify-end gap-1">
            <span v-for="scope in modelScopeLabels" :key="scope"
              class="rounded-full border border-hairline bg-canvas px-2 py-0.5 text-[10px] font-medium text-body">
              {{ scope }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="plan.features.length > 0" class="mt-6 flex-1 space-y-2">
      <div class="text-xs font-medium uppercase tracking-wide text-muted-soft">
        {{ t('payment.planFeatures') }}
      </div>
      <ul class="space-y-2">
        <li v-for="feature in plan.features" :key="feature" class="flex items-start gap-2">
          <svg class="mt-0.5 h-4 w-4 flex-shrink-0 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
          <span class="text-sm leading-relaxed text-body">{{ feature }}</span>
        </li>
      </ul>
    </div>

    <div class="mt-auto pt-6">
      <button
        type="button"
        class="btn btn-primary w-full"
        @click="emit('select', plan)"
      >
        {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import {
  formatPaymentAmount,
  normalizePaymentCurrency,
} from '@/components/payment/currency'
import {
  platformLabel,
} from '@/utils/platformColors'

const props = defineProps<{
  plan: SubscriptionPlan
  activeSubscriptions?: UserSubscription[]
  currency?: string | null
  locale?: string
}>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

const paymentCurrency = computed(() => normalizePaymentCurrency(props.currency))
const priceDisplay = computed(() =>
  formatPaymentAmount(
    Number(props.plan.price) || 0,
    paymentCurrency.value,
    props.locale,
  ),
)
const originalPriceDisplay = computed(() =>
  props.plan.original_price
    ? formatPaymentAmount(
        Number(props.plan.original_price) || 0,
        paymentCurrency.value,
        props.locale,
      )
    : '',
)

const platform = computed(() => props.plan.group_platform || '')
const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.group_id === props.plan.group_id && s.status === 'active') ?? false
)

const pLabel = computed(() => platformLabel(platform.value))

const discountText = computed(() => {
  if (!props.plan.original_price || props.plan.original_price <= 0) return ''
  const pct = Math.round((1 - props.plan.price / props.plan.original_price) * 100)
  return pct > 0 ? `-${pct}%` : ''
})

const rateDisplay = computed(() => {
  const rate = props.plan.rate_multiplier ?? 1
  return `×${Number(rate.toPrecision(10))}`
})

const MODEL_SCOPE_LABELS: Record<string, string> = {
  claude: 'Claude',
  gemini_text: 'Gemini',
  gemini_image: 'Imagen',
}

const modelScopeLabels = computed(() => {
  if (platform.value !== 'antigravity') return []
  const scopes = props.plan.supported_model_scopes
  if (!scopes || scopes.length === 0) return []
  return scopes.map(s => MODEL_SCOPE_LABELS[s] || s)
})

const validitySuffix = computed(() => {
  const u = props.plan.validity_unit || 'day'
  if (u === 'month') return t('payment.perMonth')
  if (u === 'year') return t('payment.perYear')
  return `${props.plan.validity_days}${t('payment.days')}`
})
</script>
