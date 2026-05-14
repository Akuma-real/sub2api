<template>
  <div class="card p-4">
    <div class="mb-4 flex items-center justify-between gap-3">
      <h3 class="text-sm font-semibold text-ink">
        {{ t("admin.dashboard.groupDistribution") }}
      </h3>
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
    </div>
    <div v-if="loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div
      v-else-if="displayGroupStats.length > 0 && chartData"
      class="flex items-center gap-6"
    >
      <div class="h-48 w-48">
        <Doughnut :data="chartData" :options="doughnutOptions" />
      </div>
      <div class="max-h-48 flex-1 overflow-y-auto">
        <table class="w-full text-xs">
          <thead>
            <tr class="text-muted">
              <th class="pb-2 text-left">{{ t("admin.dashboard.group") }}</th>
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
            <template v-for="group in displayGroupStats" :key="group.group_id">
              <tr
                class="border-t border-hairline-soft transition-colors"
                :class="
                  group.group_id > 0
                    ? 'cursor-pointer hover:bg-surface-soft '
                    : ''
                "
                @click="
                  group.group_id > 0 && toggleBreakdown('group', group.group_id)
                "
              >
                <td
                  class="max-w-[100px] truncate py-1.5 font-medium"
                  :class="
                    group.group_id > 0
                      ? 'text-primary-700 hover:text-body-strong '
                      : 'text-ink '
                  "
                  :title="group.group_name || String(group.group_id)"
                >
                  <span class="inline-flex items-center gap-1">
                    <svg
                      v-if="
                        group.group_id > 0 &&
                        expandedKey === `group-${group.group_id}`
                      "
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
                      v-else-if="group.group_id > 0"
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
                    {{ group.group_name || t("admin.dashboard.noGroup") }}
                  </span>
                </td>
                <td class="py-1.5 text-right text-body">
                  {{ formatNumber(group.requests) }}
                </td>
                <td class="py-1.5 text-right text-body">
                  {{ formatTokens(group.total_tokens) }}
                </td>
                <td class="py-1.5 text-right text-success">
                  ${{ formatCost(group.actual_cost) }}
                </td>
                <td class="py-1.5 text-right text-warning">
                  ${{ formatCost(group.account_cost) }}
                </td>
                <td class="py-1.5 text-right text-muted-soft">
                  ${{ formatCost(group.cost) }}
                </td>
              </tr>
              <!-- User breakdown sub-rows -->
              <tr v-if="expandedKey === `group-${group.group_id}`">
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
import type { GroupStat, UserBreakdownItem } from "@/types";
import { getUserBreakdown } from "@/api/admin/dashboard";

ChartJS.register(ArcElement, Tooltip, Legend);

const { t } = useI18n();

type DistributionMetric = "tokens" | "actual_cost";

const props = withDefaults(
  defineProps<{
    groupStats: GroupStat[];
    loading?: boolean;
    metric?: DistributionMetric;
    showMetricToggle?: boolean;
    startDate?: string;
    endDate?: string;
    filters?: Record<string, any>;
  }>(),
  {
    loading: false,
    metric: "tokens",
    showMetricToggle: false,
  },
);

const emit = defineEmits<{
  "update:metric": [value: DistributionMetric];
}>();

const expandedKey = ref<string | null>(null);
const breakdownItems = ref<UserBreakdownItem[]>([]);
const breakdownLoading = ref(false);

const toggleBreakdown = async (type: string, id: number | string) => {
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
      group_id: Number(id),
    });
    breakdownItems.value = res.users || [];
  } catch {
    breakdownItems.value = [];
  } finally {
    breakdownLoading.value = false;
  }
};

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
];

const displayGroupStats = computed(() => {
  if (!props.groupStats?.length) return [];

  const metricKey =
    props.metric === "actual_cost" ? "actual_cost" : "total_tokens";
  return [...props.groupStats].sort((a, b) => b[metricKey] - a[metricKey]);
});

const chartData = computed(() => {
  if (!props.groupStats?.length) return null;

  return {
    labels: displayGroupStats.value.map(
      (g) => g.group_name || String(g.group_id),
    ),
    datasets: [
      {
        data: displayGroupStats.value.map((g) =>
          props.metric === "actual_cost" ? g.actual_cost : g.total_tokens,
        ),
        backgroundColor: chartColors.slice(0, displayGroupStats.value.length),
        borderWidth: 0,
      },
    ],
  };
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
