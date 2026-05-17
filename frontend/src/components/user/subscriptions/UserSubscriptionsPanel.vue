<template>
  <div class="space-y-6">
    <div v-if="loading" class="flex justify-center py-12">
      <div
        class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
      ></div>
    </div>

    <div v-else-if="subscriptions.length === 0" class="card p-12 text-center">
      <div
        class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-surface-card"
      >
        <Icon name="creditCard" size="xl" class="text-muted-soft" />
      </div>
      <h3 class="mb-2 text-lg font-semibold text-ink">
        {{ t("userSubscriptions.noActiveSubscriptions") }}
      </h3>
      <p class="text-muted">
        {{ t("userSubscriptions.noActiveSubscriptionsDesc") }}
      </p>
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-2">
      <div
        v-for="subscription in subscriptions"
        :key="subscription.id"
        class="overflow-hidden rounded-lg border bg-canvas"
        :class="platformBorderClass(subscription.group?.platform || '')"
      >
        <div
          class="flex items-center justify-between border-b border-hairline-soft p-4"
        >
          <div class="flex min-w-0 items-center gap-3">
            <div
              :class="[
                'h-1.5 w-1.5 shrink-0 rounded-full',
                platformAccentDotClass(subscription.group?.platform || ''),
              ]"
            />
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <h3 class="truncate font-semibold text-ink">
                  {{
                    subscription.group?.name || `Group #${subscription.group_id}`
                  }}
                </h3>
                <span
                  :class="[
                    'rounded-md border px-2 py-0.5 text-[11px] font-medium',
                    platformBadgeClass(subscription.group?.platform || ''),
                  ]"
                >
                  {{ platformLabel(subscription.group?.platform || "") }}
                </span>
              </div>
              <p
                v-if="subscription.group?.description"
                class="mt-0.5 truncate text-xs text-muted"
              >
                {{ subscription.group.description }}
              </p>
            </div>
          </div>
          <div class="flex shrink-0 items-center gap-2">
            <span
              :class="[
                'rounded-full px-2 py-0.5 text-xs font-medium',
                subscription.status === 'active'
                  ? 'bg-success/15 text-success '
                  : subscription.status === 'expired'
                    ? 'bg-surface-card text-body '
                    : 'bg-error/15 text-error ',
              ]"
            >
              {{ t(`userSubscriptions.status.${subscription.status}`) }}
            </span>
            <button
              v-if="subscription.status === 'active'"
              :class="[
                'rounded-lg px-3 py-1.5 text-xs font-semibold text-on-primary transition-colors',
                platformButtonClass(subscription.group?.platform || ''),
              ]"
              @click="
                router.push({
                  path: '/purchase',
                  query: {
                    tab: 'subscription',
                    group: String(subscription.group_id),
                  },
                })
              "
            >
              {{ t("payment.renewNow") }}
            </button>
          </div>
        </div>

        <div class="space-y-4 p-4">
          <div
            v-if="subscription.expires_at"
            class="flex items-center justify-between text-sm"
          >
            <span class="text-muted">{{ t("userSubscriptions.expires") }}</span>
            <span :class="getExpirationClass(subscription.expires_at)">
              {{ formatExpirationDate(subscription.expires_at) }}
            </span>
          </div>
          <div v-else class="flex items-center justify-between text-sm">
            <span class="text-muted">{{ t("userSubscriptions.expires") }}</span>
            <span class="text-body">{{
              t("userSubscriptions.noExpiration")
            }}</span>
          </div>

          <UsageLimitRow
            v-if="subscription.group?.daily_limit_usd"
            :label="t('userSubscriptions.daily')"
            :used="subscription.daily_usage_usd"
            :limit="subscription.group.daily_limit_usd"
            :reset="
              subscription.daily_window_start
                ? t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.daily_window_start, 24),
                  })
                : ''
            "
          />

          <UsageLimitRow
            v-if="subscription.group?.weekly_limit_usd"
            :label="t('userSubscriptions.weekly')"
            :used="subscription.weekly_usage_usd"
            :limit="subscription.group.weekly_limit_usd"
            :reset="
              subscription.weekly_window_start
                ? t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.weekly_window_start, 168),
                  })
                : ''
            "
          />

          <UsageLimitRow
            v-if="subscription.group?.monthly_limit_usd"
            :label="t('userSubscriptions.monthly')"
            :used="subscription.monthly_usage_usd"
            :limit="subscription.group.monthly_limit_usd"
            :reset="
              subscription.monthly_window_start
                ? t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.monthly_window_start, 720),
                  })
                : ''
            "
          />

          <div
            v-if="
              !subscription.group?.daily_limit_usd &&
              !subscription.group?.weekly_limit_usd &&
              !subscription.group?.monthly_limit_usd
            "
            class="flex items-center justify-center rounded-xl bg-surface-soft py-6"
          >
            <div class="flex items-center gap-3">
              <span class="text-4xl text-success">∞</span>
              <div>
                <p class="text-sm font-medium text-success">
                  {{ t("userSubscriptions.unlimited") }}
                </p>
                <p class="text-xs text-success/70">
                  {{ t("userSubscriptions.unlimitedDesc") }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineComponent, h, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import subscriptionsAPI from "@/api/subscriptions";
import Icon from "@/components/icons/Icon.vue";
import { useAppStore } from "@/stores/app";
import type { UserSubscription } from "@/types";
import { formatDateOnly } from "@/utils/format";
import {
  platformBadgeClass,
  platformBorderClass,
  platformButtonClass,
  platformLabel,
} from "@/utils/platformColors";

const UsageLimitRow = defineComponent({
  props: {
    label: { type: String, required: true },
    used: { type: Number, default: 0 },
    limit: { type: Number, required: true },
    reset: { type: String, default: "" },
  },
  setup(props) {
    const width = () =>
      props.limit ? `${Math.min(((props.used || 0) / props.limit) * 100, 100)}%` : "0%";
    const barClass = () => {
      if (!props.limit) return "bg-muted-soft";
      const percentage = ((props.used || 0) / props.limit) * 100;
      if (percentage >= 90) return "bg-error";
      if (percentage >= 70) return "bg-accent-amber";
      return "bg-success";
    };
    return () =>
      h("div", { class: "space-y-2" }, [
        h("div", { class: "flex items-center justify-between" }, [
          h("span", { class: "text-sm font-medium text-body" }, props.label),
          h(
            "span",
            { class: "text-sm text-muted" },
            `$${(props.used || 0).toFixed(2)} / $${props.limit.toFixed(2)}`,
          ),
        ]),
        h("div", { class: "relative h-2 overflow-hidden rounded-full bg-hairline" }, [
          h("div", {
            class: ["absolute inset-y-0 left-0 rounded-full transition-all duration-300", barClass()],
            style: { width: width() },
          }),
        ]),
        props.reset ? h("p", { class: "text-xs text-muted" }, props.reset) : null,
      ]);
  },
});

const { t } = useI18n();
const router = useRouter();
const appStore = useAppStore();

const subscriptions = ref<UserSubscription[]>([]);
const loading = ref(true);

function platformAccentDotClass(platform: string): string {
  switch (platform) {
    case "anthropic":
      return "bg-accent-amber";
    case "openai":
      return "bg-success";
    case "antigravity":
      return "bg-primary-500";
    case "gemini":
      return "bg-accent-teal";
    default:
      return "bg-muted-soft";
  }
}

async function loadSubscriptions() {
  try {
    loading.value = true;
    subscriptions.value = await subscriptionsAPI.getMySubscriptions();
  } catch (error) {
    console.error("Failed to load subscriptions: ", error);
    appStore.showError(t("userSubscriptions.failedToLoad"));
  } finally {
    loading.value = false;
  }
}

function formatExpirationDate(expiresAt: string): string {
  const now = new Date();
  const expires = new Date(expiresAt);
  const diff = expires.getTime() - now.getTime();
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24));

  if (days < 0) return t("userSubscriptions.status.expired");

  const dateStr = formatDateOnly(expires);
  if (days === 0) return `${dateStr} (${t("common.today")})`;
  if (days === 1) return `${dateStr} (${t("common.tomorrow")})`;
  return t("userSubscriptions.daysRemaining", { days }) + ` (${dateStr})`;
}

function getExpirationClass(expiresAt: string): string {
  const days = Math.ceil(
    (new Date(expiresAt).getTime() - Date.now()) / (1000 * 60 * 60 * 24),
  );
  if (days <= 0) return "text-error font-medium";
  if (days <= 3) return "text-error ";
  if (days <= 7) return "text-warning ";
  return "text-body ";
}

function formatResetTime(windowStart: string | null, windowHours: number): string {
  if (!windowStart) return t("userSubscriptions.windowNotActive");
  const end = new Date(windowStart).getTime() + windowHours * 60 * 60 * 1000;
  const diff = end - Date.now();
  if (diff <= 0) return t("userSubscriptions.windowNotActive");
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  if (hours > 24) return `${Math.floor(hours / 24)}d ${hours % 24}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}

onMounted(loadSubscriptions);
</script>
