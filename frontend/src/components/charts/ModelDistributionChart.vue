<template>
  <div class="card p-4">
    <div class="mb-4 flex items-center justify-between gap-3">
      <h3 class="text-sm font-semibold text-ink">
        {{
          !enableRankingView || activeView === "model_distribution"
            ? t("admin.dashboard.modelDistribution")
            : t("admin.dashboard.spendingRankingTitle")
        }}
      </h3>
      <div class="flex flex-wrap items-center justify-end gap-2">
        <div
          v-if="showSourceToggle"
          class="inline-flex rounded-lg border border-hairline bg-surface-soft p-0.5"
        >
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="
              source === 'requested'
                ? 'bg-canvas text-ink shadow-sm '
                : 'text-muted hover:text-body '
            "
            @click="emit('update:source', 'requested')"
          >
            {{ t("usage.requestedModel") }}
          </button>
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="
              source === 'upstream'
                ? 'bg-canvas text-ink shadow-sm '
                : 'text-muted hover:text-body '
            "
            @click="emit('update:source', 'upstream')"
          >
            {{ t("usage.upstreamModel") }}
          </button>
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="
              source === 'mapping'
                ? 'bg-canvas text-ink shadow-sm '
                : 'text-muted hover:text-body '
            "
            @click="emit('update:source', 'mapping')"
          >
            {{ t("usage.mapping") }}
          </button>
        </div>
        <div
          v-if="showMetricToggle"
          class="inline-flex rounded-lg border border-hairline bg-surface-soft p-0.5"
        >
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="
              metric === 'tokens'
                ? 'bg-canvas text-ink shadow-sm '
                : 'text-muted hover:text-body '
            "
            @click="emit('update:metric', 'tokens')"
          >
            {{ t("admin.dashboard.metricTokens") }}
          </button>
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="
              metric === 'actual_cost'
                ? 'bg-canvas text-ink shadow-sm '
                : 'text-muted hover:text-body '
            "
            @click="emit('update:metric', 'actual_cost')"
          >
            {{ t("admin.dashboard.metricActualCost") }}
          </button>
        </div>
        <div
          v-if="enableRankingView"
          class="inline-flex rounded-lg bg-surface-card p-1"
        >
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="
              activeView === 'model_distribution'
                ? 'bg-canvas text-ink shadow-sm '
                : 'text-muted hover:text-body '
            "
            @click="activeView = 'model_distribution'"
          >
            {{ t("admin.dashboard.viewModelDistribution") }}
          </button>
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="
              activeView === 'spending_ranking'
                ? 'bg-canvas text-ink shadow-sm '
                : 'text-muted hover:text-body '
            "
            @click="activeView = 'spending_ranking'"
          >
            {{ t("admin.dashboard.viewSpendingRanking") }}
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="activeView === 'model_distribution' && loading"
      class="flex h-48 items-center justify-center"
    >
      <LoadingSpinner />
    </div>
    <div
      v-else-if="
        activeView === 'model_distribution' &&
        displayModelStats.length > 0 &&
        chartData
      "
      class="flex items-center gap-6"
    >
      <div class="h-48 w-48">
        <Doughnut :data="chartData" :options="doughnutOptions" />
      </div>
      <div class="max-h-48 flex-1 overflow-y-auto">
        <table class="w-full text-xs">
          <thead>
            <tr class="text-muted">
              <th class="pb-2 text-left">{{ t("admin.dashboard.model") }}</th>
              <th class="pb-2 text-right">
                {{ t("admin.dashboard.requests") }}
              </th>
              <th class="pb-2 text-right">{{ t("admin.dashboard.tokens") }}</th>
              <th class="pb-2 text-right">{{ t("admin.dashboard.actual") }}</th>
              <th class="pb-2 text-right">
                {{ t("admin.dashboard.accountCost") }}
              </th>
              <th class="pb-2 text-right">
                {{ t("admin.dashboard.standard") }}
              </th>
            </tr>
          </thead>
          <tbody>
            <template v-for="model in displayModelStats" :key="model.model">
              <tr
                class="border-t border-hairline-soft cursor-pointer transition-colors hover:bg-surface-soft"
                @click="toggleBreakdown('model', model.model)"
              >
                <td
                  class="max-w-[100px] truncate py-1.5 font-medium text-primary-700 hover:text-body-strong"
                  :title="model.model"
                >
                  <span class="inline-flex items-center gap-1">
                    <svg
                      v-if="expandedKey === `model-${model.model}`"
                      class="h-3 w-3 shrink-0"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M19 9l-7 7-7-7"
                      />
                    </svg>
                    <svg
                      v-else
                      class="h-3 w-3 shrink-0"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M9 5l7 7-7 7"
                      />
                    </svg>
                    {{ model.model }}
                  </span>
                </td>
                <td class="py-1.5 text-right text-body">
                  {{ formatNumber(model.requests) }}
                </td>
                <td class="py-1.5 text-right text-body">
                  {{ formatTokens(model.total_tokens) }}
                </td>
                <td class="py-1.5 text-right text-success">
                  ${{ formatCost(model.actual_cost) }}
                </td>
                <td class="py-1.5 text-right text-warning">
                  ${{ formatCost(model.account_cost) }}
                </td>
                <td class="py-1.5 text-right text-muted-soft">
                  ${{ formatCost(model.cost) }}
                </td>
              </tr>
              <tr v-if="expandedKey === `model-${model.model}`">
                <td colspan="6" class="p-0">
                  <UserBreakdownSubTable
                    :items="breakdownItems"
                    :loading="breakdownLoading"
                  />
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
    <div
      v-else-if="activeView === 'model_distribution'"
      class="flex h-48 items-center justify-center text-sm text-muted"
    >
      {{ t("admin.dashboard.noDataAvailable") }}
    </div>

    <div
      v-else-if="rankingLoading"
      class="flex h-48 items-center justify-center"
    >
      <LoadingSpinner />
    </div>
    <div
      v-else-if="rankingError"
      class="flex h-48 items-center justify-center text-sm text-muted"
    >
      {{ t("admin.dashboard.failedToLoad") }}
    </div>
    <div
      v-else-if="rankingDisplayItems.length > 0 && rankingChartData"
      class="flex items-center gap-6"
    >
      <div class="h-48 w-48">
        <Doughnut :data="rankingChartData" :options="rankingDoughnutOptions" />
      </div>
      <div class="max-h-48 flex-1 overflow-y-auto">
        <table class="w-full text-xs">
          <thead>
            <tr class="text-muted">
              <th class="pb-2 text-left">
                {{ t("admin.dashboard.spendingRankingUser") }}
              </th>
              <th class="pb-2 text-right">
                {{ t("admin.dashboard.spendingRankingRequests") }}
              </th>
              <th class="pb-2 text-right">
                {{ t("admin.dashboard.spendingRankingTokens") }}
              </th>
              <th class="pb-2 text-right">
                {{ t("admin.dashboard.spendingRankingSpend") }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(item, index) in rankingDisplayItems"
              :key="item.isOther ? 'others' : `${item.user_id}-${index}`"
              class="border-t border-hairline-soft transition-colors"
              :class="
                item.isOther
                  ? 'bg-surface-soft/70 '
                  : 'cursor-pointer hover:bg-surface-soft '
              "
              @click="item.isOther ? undefined : emit('ranking-click', item)"
            >
              <td class="py-1.5">
                <div class="flex min-w-0 items-center gap-2">
                  <span class="shrink-0 text-[11px] font-semibold text-muted">
                    {{ item.isOther ? "Σ" : `#${index + 1}` }}
                  </span>
                  <span
                    class="block max-w-[140px] truncate font-medium text-ink"
                    :title="getRankingRowLabel(item)"
                  >
                    {{ getRankingRowLabel(item) }}
                  </span>
                </div>
              </td>
              <td class="py-1.5 text-right text-body">
                {{ formatNumber(item.requests) }}
              </td>
              <td class="py-1.5 text-right text-body">
                {{ formatTokens(item.tokens) }}
              </td>
              <td class="py-1.5 text-right text-success">
                ${{ formatCost(item.actual_cost) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div
      v-else
      class="flex h-48 items-center justify-center text-sm text-muted"
    >
      {{ t("admin.dashboard.noDataAvailable") }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from "chart.js";
import { Doughnut } from "vue-chartjs";
import LoadingSpinner from "@/components/common/LoadingSpinner.vue";
import UserBreakdownSubTable from "./UserBreakdownSubTable.vue";
import type {
  ModelStat,
  UserSpendingRankingItem,
  UserBreakdownItem,
} from "@/types";
import { getUserBreakdown } from "@/api/admin/dashboard";

ChartJS.register(ArcElement, Tooltip, Legend);

const { t } = useI18n();

type DistributionMetric = "tokens" | "actual_cost";
type ModelSource = "requested" | "upstream" | "mapping";
type RankingDisplayItem = UserSpendingRankingItem & { isOther?: boolean };
const props = withDefaults(
  defineProps<{
    modelStats: ModelStat[];
    upstreamModelStats?: ModelStat[];
    mappingModelStats?: ModelStat[];
    source?: ModelSource;
    enableRankingView?: boolean;
    rankingItems?: UserSpendingRankingItem[];
    rankingTotalActualCost?: number;
    rankingTotalRequests?: number;
    rankingTotalTokens?: number;
    loading?: boolean;
    metric?: DistributionMetric;
    showSourceToggle?: boolean;
    showMetricToggle?: boolean;
    rankingLoading?: boolean;
    rankingError?: boolean;
    startDate?: string;
    endDate?: string;
    filters?: Record<string, any>;
  }>(),
  {
    upstreamModelStats: () => [],
    mappingModelStats: () => [],
    source: "requested",
    enableRankingView: false,
    rankingItems: () => [],
    rankingTotalActualCost: 0,
    rankingTotalRequests: 0,
    rankingTotalTokens: 0,
    loading: false,
    metric: "tokens",
    showSourceToggle: false,
    showMetricToggle: false,
    rankingLoading: false,
    rankingError: false,
  },
);

const expandedKey = ref<string | null>(null);
const breakdownItems = ref<UserBreakdownItem[]>([]);
const breakdownLoading = ref(false);

const toggleBreakdown = async (type: string, id: string) => {
  const key = `${type}-${id}`;
  if (expandedKey.value === key) {
    expandedKey.value = null;
    return;
  }
  expandedKey.value = key;
  breakdownLoading.value = true;
  breakdownItems.value = [];
  try {
    const res = await getUserBreakdown({
      ...props.filters,
      start_date: props.startDate,
      end_date: props.endDate,
      model: id,
      model_source: props.source,
    });
    breakdownItems.value = res.users || [];
  } catch {
    breakdownItems.value = [];
  } finally {
    breakdownLoading.value = false;
  }
};

const emit = defineEmits<{
  "update:metric": [value: DistributionMetric];
  "update:source": [value: ModelSource];
  "ranking-click": [item: UserSpendingRankingItem];
}>();

const enableRankingView = computed(() => props.enableRankingView);
const activeView = ref<"model_distribution" | "spending_ranking">(
  "model_distribution",
);

const chartColors = [
  "#cc785c",
  "#5db872",
  "#e8a55a",
  "#c64545",
  "#cc785c",
  "#cc785c",
  "#5db8a6",
  "#e8a55a",
  "#cc785c",
  "#5db872",
  "#5db8a6",
  "#cc785c",
];

const displayModelStats = computed(() => {
  const sourceStats =
    props.source === "upstream"
      ? props.upstreamModelStats
      : props.source === "mapping"
        ? props.mappingModelStats
        : props.modelStats;
  if (!sourceStats?.length) return [];

  const metricKey =
    props.metric === "actual_cost" ? "actual_cost" : "total_tokens";
  return [...sourceStats].sort((a, b) => b[metricKey] - a[metricKey]);
});

const chartData = computed(() => {
  if (!displayModelStats.value.length) return null;

  return {
    labels: displayModelStats.value.map((m) => m.model),
    datasets: [
      {
        data: displayModelStats.value.map((m) =>
          props.metric === "actual_cost" ? m.actual_cost : m.total_tokens,
        ),
        backgroundColor: chartColors.slice(0, displayModelStats.value.length),
        borderWidth: 0,
      },
    ],
  };
});

const rankingChartData = computed(() => {
  if (!props.rankingItems?.length) return null;

  const labels = props.rankingItems.map(
    (item, index) => `#${index + 1} ${getRankingUserLabel(item)}`,
  );
  const data = props.rankingItems.map((item) => item.actual_cost);
  const backgroundColor = chartColors.slice(0, props.rankingItems.length);

  if (otherRankingItem.value) {
    labels.push(t("admin.dashboard.spendingRankingOther"));
    data.push(otherRankingItem.value.actual_cost);
    backgroundColor.push("#8e8b82");
  }

  return {
    labels,
    datasets: [
      {
        data,
        backgroundColor,
        borderWidth: 0,
      },
    ],
  };
});

const otherRankingItem = computed<RankingDisplayItem | null>(() => {
  if (!props.rankingItems?.length) return null;

  const rankedActualCost = props.rankingItems.reduce(
    (sum, item) => sum + item.actual_cost,
    0,
  );
  const rankedRequests = props.rankingItems.reduce(
    (sum, item) => sum + item.requests,
    0,
  );
  const rankedTokens = props.rankingItems.reduce(
    (sum, item) => sum + item.tokens,
    0,
  );

  const otherActualCost = Math.max(
    (props.rankingTotalActualCost || 0) - rankedActualCost,
    0,
  );
  const otherRequests = Math.max(
    (props.rankingTotalRequests || 0) - rankedRequests,
    0,
  );
  const otherTokens = Math.max(
    (props.rankingTotalTokens || 0) - rankedTokens,
    0,
  );

  if (otherActualCost <= 0.000001 && otherRequests <= 0 && otherTokens <= 0)
    return null;

  return {
    user_id: 0,
    email: "",
    actual_cost: otherActualCost,
    requests: otherRequests,
    tokens: otherTokens,
    isOther: true,
  };
});

const rankingDisplayItems = computed<RankingDisplayItem[]>(() => {
  if (!props.rankingItems?.length) return [];
  return otherRankingItem.value
    ? [...props.rankingItems, otherRankingItem.value]
    : [...props.rankingItems];
});

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const value = context.raw as number;
          const total = context.dataset.data.reduce(
            (a: number, b: number) => a + b,
            0,
          );
          const percentage =
            total > 0 ? ((value / total) * 100).toFixed(1) : "0.0";
          const formattedValue =
            props.metric === "actual_cost"
              ? `$${formatCost(value)}`
              : formatTokens(value);
          return `${context.label}: ${formattedValue} (${percentage}%)`;
        },
      },
    },
  },
}));

const rankingDoughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const value = context.raw as number;
          const total = context.dataset.data.reduce(
            (a: number, b: number) => a + b,
            0,
          );
          const percentage =
            total > 0 ? ((value / total) * 100).toFixed(1) : "0.0";
          return `${context.label}: $${formatCost(value)} (${percentage}%)`;
        },
      },
    },
  },
}));

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`;
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`;
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`;
  }
  return value.toLocaleString();
};

const formatNumber = (value: number): string => {
  return value.toLocaleString();
};

const getRankingUserLabel = (item: UserSpendingRankingItem): string => {
  if (item.email) return item.email;
  return t("admin.redeem.userPrefix", { id: item.user_id });
};

const getRankingRowLabel = (item: RankingDisplayItem): string => {
  if (item.isOther) return t("admin.dashboard.spendingRankingOther");
  return getRankingUserLabel(item);
};

const formatCost = (value?: number): string => {
  const amount = Number(value) || 0;
  if (amount >= 1000) {
    return (amount / 1000).toFixed(2) + "K";
  } else if (amount >= 1) {
    return amount.toFixed(2);
  } else if (amount >= 0.01) {
    return amount.toFixed(3);
  }
  return amount.toFixed(4);
};
</script>
