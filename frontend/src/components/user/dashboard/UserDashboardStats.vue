<template>
  <!-- Row 1: Core Stats -->
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
    <!-- Balance -->
    <div v-if="!isSimple" class="dashboard-stat-card card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-success/15 p-2">
          <svg
            class="h-5 w-5 text-success"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
            />
          </svg>
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-muted">
            {{ t("dashboard.balance") }}
          </p>
          <p class="truncate text-xl font-bold text-success" :title="`$${formatBalance(balance)}`">
            ${{ formatBalance(balance) }}
          </p>
          <p class="text-xs text-muted">{{ t("common.available") }}</p>
        </div>
      </div>
    </div>

    <!-- API Keys -->
    <div class="dashboard-stat-card card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-accent-teal/15 p-2">
          <Icon
            name="key"
            size="md"
            class="text-primary-700"
            :stroke-width="2"
          />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-muted">
            {{ t("dashboard.apiKeys") }}
          </p>
          <p class="truncate text-xl font-bold text-ink" :title="String(stats?.total_api_keys || 0)">
            {{ stats?.total_api_keys || 0 }}
          </p>
          <p class="text-xs text-success">
            {{ stats?.active_api_keys || 0 }} {{ t("common.active") }}
          </p>
        </div>
      </div>
    </div>

    <!-- Today Requests -->
    <div class="dashboard-stat-card card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-success/15 p-2">
          <Icon
            name="chart"
            size="md"
            class="text-success"
            :stroke-width="2"
          />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-muted">
            {{ t("dashboard.todayRequests") }}
          </p>
          <p class="truncate text-xl font-bold text-ink" :title="String(stats?.today_requests || 0)">
            {{ stats?.today_requests || 0 }}
          </p>
          <p class="truncate text-xs text-muted">
            {{ t("common.total") }}:
            {{ formatNumber(stats?.total_requests || 0) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Today Cost -->
    <div class="dashboard-stat-card card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-primary-100 p-2">
          <Icon
            name="dollar"
            size="md"
            class="text-primary-700"
            :stroke-width="2"
          />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-muted">
            {{ t("dashboard.todayCost") }}
          </p>
          <p class="truncate text-xl font-bold text-ink">
            <span class="text-primary-700" :title="t('dashboard.actual')"
              >${{ formatCost(stats?.today_actual_cost || 0) }}</span
            >
            <span
              class="text-sm font-normal text-muted-soft"
              :title="t('dashboard.standard')"
            >
              / ${{ formatCost(stats?.today_cost || 0) }}</span
            >
          </p>
          <p class="truncate text-xs">
            <span class="text-muted">{{ t("common.total") }}: </span>
            <span class="text-primary-700" :title="t('dashboard.actual')"
              >${{ formatCost(stats?.total_actual_cost || 0) }}</span
            >
            <span class="text-muted-soft" :title="t('dashboard.standard')">
              / ${{ formatCost(stats?.total_cost || 0) }}</span
            >
          </p>
        </div>
      </div>
    </div>
  </div>

  <!-- Row 2: Token Stats -->
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
    <!-- Today Tokens -->
    <div class="dashboard-stat-card card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-accent-amber/15 p-2">
          <Icon
            name="cube"
            size="md"
            class="text-warning"
            :stroke-width="2"
          />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-muted">
            {{ t("dashboard.todayTokens") }}
          </p>
          <p class="truncate text-xl font-bold text-ink" :title="formatTokens(stats?.today_tokens || 0)">
            {{ formatTokens(stats?.today_tokens || 0) }}
          </p>
          <p class="truncate text-xs text-muted">
            {{ t("dashboard.input") }}:
            {{ formatTokens(stats?.today_input_tokens || 0) }} /
            {{ t("dashboard.output") }}:
            {{ formatTokens(stats?.today_output_tokens || 0) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Total Tokens -->
    <div class="dashboard-stat-card card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-primary-100 p-2">
          <Icon
            name="database"
            size="md"
            class="text-primary-600"
            :stroke-width="2"
          />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-muted">
            {{ t("dashboard.totalTokens") }}
          </p>
          <p class="truncate text-xl font-bold text-ink" :title="formatTokens(stats?.total_tokens || 0)">
            {{ formatTokens(stats?.total_tokens || 0) }}
          </p>
          <p class="truncate text-xs text-muted">
            {{ t("dashboard.input") }}:
            {{ formatTokens(stats?.total_input_tokens || 0) }} /
            {{ t("dashboard.output") }}:
            {{ formatTokens(stats?.total_output_tokens || 0) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Performance (RPM/TPM) -->
    <div class="dashboard-stat-card card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-primary-100 p-2">
          <Icon
            name="bolt"
            size="md"
            class="text-primary-600"
            :stroke-width="2"
          />
        </div>
        <div class="flex-1">
          <p class="text-xs font-medium text-muted">
            {{ t("dashboard.performance") }}
          </p>
          <div class="flex items-baseline gap-2">
            <p class="text-xl font-bold text-ink">
              {{ formatTokens(stats?.rpm || 0) }}
            </p>
            <span class="text-xs text-muted">RPM</span>
          </div>
          <div class="flex items-baseline gap-2">
            <p class="text-sm font-semibold text-primary-600">
              {{ formatTokens(stats?.tpm || 0) }}
            </p>
            <span class="text-xs text-muted">TPM</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Avg Response Time -->
    <div class="dashboard-stat-card card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-error/15 p-2">
          <Icon
            name="clock"
            size="md"
            class="text-error"
            :stroke-width="2"
          />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-muted">
            {{ t("dashboard.avgResponse") }}
          </p>
          <p class="truncate text-xl font-bold text-ink" :title="formatDuration(stats?.average_duration_ms || 0)">
            {{ formatDuration(stats?.average_duration_ms || 0) }}
          </p>
          <p class="text-xs text-muted">{{ t("dashboard.averageTime") }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";
import type { UserDashboardStats as UserStatsType } from "@/api/usage";

defineProps<{
  stats: UserStatsType;
  balance: number;
  isSimple: boolean;
}>();
const { t } = useI18n();

const formatBalance = (b: number) =>
  new Intl.NumberFormat("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(b);

const formatNumber = (n: number) => n.toLocaleString();
const formatCost = (c: number) => c.toFixed(4);
const formatTokens = (t: number) => {
  if (t >= 1_000_000) return `${(t / 1_000_000).toFixed(1)}M`;
  if (t >= 1000) return `${(t / 1000).toFixed(1)}K`;
  return t.toString();
};
const formatDuration = (ms: number) =>
  ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`;
</script>

<style scoped>
.dashboard-stat-card {
  min-width: 0;
}

.dashboard-stat-card > div {
  min-width: 0;
}

.dashboard-stat-card svg {
  flex-shrink: 0;
}
</style>
