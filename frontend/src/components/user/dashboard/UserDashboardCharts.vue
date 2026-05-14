<template>
  <div class="space-y-6">
    <!-- Date Range Filter -->
    <div class="card p-4">
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex items-center gap-2">
          <span class="text-sm font-medium text-body"
            >{{ t("dashboard.timeRange") }}:</span
          >
          <DateRangePicker
            :start-date="startDate"
            :end-date="endDate"
            @update:startDate="$emit('update:startDate', $event)"
            @update:endDate="$emit('update:endDate', $event)"
            @change="$emit('dateRangeChange', $event)"
          />
        </div>
        <button
          @click="$emit('refresh')"
          :disabled="loading"
          class="btn btn-secondary"
        >
          {{ t("common.refresh") }}
        </button>
        <div class="ml-auto flex items-center gap-2">
          <span class="text-sm font-medium text-body"
            >{{ t("dashboard.granularity") }}:</span
          >
          <div class="w-28">
            <Select
              :model-value="granularity"
              :options="[
                { value: 'day', label: t('dashboard.day') },
                { value: 'hour', label: t('dashboard.hour') },
              ]"
              @update:model-value="$emit('update:granularity', $event)"
              @change="$emit('granularityChange')"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Grid -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- Model Distribution Chart -->
      <div class="card relative overflow-hidden p-4">
        <div
          v-if="loading"
          class="absolute inset-0 z-10 flex items-center justify-center bg-canvas/50 "
        >
          <LoadingSpinner size="md" />
        </div>
        <h3 class="mb-4 text-sm font-semibold text-ink">
          {{ t("dashboard.modelDistribution") }}
        </h3>
        <div class="flex items-center gap-6">
          <div class="h-48 w-48">
            <Doughnut
              v-if="modelData"
              :data="modelData"
              :options="doughnutOptions"
            />
            <div
              v-else
              class="flex h-full items-center justify-center text-sm text-muted"
            >
              {{ t("dashboard.noDataAvailable") }}
            </div>
          </div>
          <div class="max-h-48 flex-1 overflow-y-auto">
            <table class="w-full text-xs">
              <thead>
                <tr class="text-muted">
                  <th class="pb-2 text-left">{{ t("dashboard.model") }}</th>
                  <th class="pb-2 text-right">{{ t("dashboard.requests") }}</th>
                  <th class="pb-2 text-right">{{ t("dashboard.tokens") }}</th>
                  <th class="pb-2 text-right">{{ t("dashboard.actual") }}</th>
                  <th class="pb-2 text-right">{{ t("dashboard.standard") }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="model in models"
                  :key="model.model"
                  class="border-t border-hairline-soft"
                >
                  <td
                    class="max-w-[100px] truncate py-1.5 font-medium text-ink"
                    :title="model.model"
                  >
                    {{ model.model }}
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
                  <td class="py-1.5 text-right text-muted-soft">
                    ${{ formatCost(model.cost) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Token Usage Trend Chart -->
      <TokenUsageTrend :trend-data="trend" :loading="loading" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import LoadingSpinner from "@/components/common/LoadingSpinner.vue";
import DateRangePicker from "@/components/common/DateRangePicker.vue";
import Select from "@/components/common/Select.vue";
import { Doughnut } from "vue-chartjs";
import TokenUsageTrend from "@/components/charts/TokenUsageTrend.vue";
import type { TrendDataPoint, ModelStat } from "@/types";
import {
  formatCostFixed as formatCost,
  formatNumberLocaleString as formatNumber,
  formatTokensK as formatTokens,
} from "@/utils/format";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from "chart.js";
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler,
);

const props = defineProps<{
  loading: boolean;
  startDate: string;
  endDate: string;
  granularity: string;
  trend: TrendDataPoint[];
  models: ModelStat[];
}>();
defineEmits([
  "update:startDate",
  "update:endDate",
  "update:granularity",
  "dateRangeChange",
  "granularityChange",
  "refresh",
]);
const { t } = useI18n();

const modelData = computed(() =>
  !props.models?.length
    ? null
    : {
        labels: props.models.map((m: ModelStat) => m.model),
        datasets: [
          {
            data: props.models.map((m: ModelStat) => m.total_tokens),
            backgroundColor: [
              "#cc785c",
              "#5db872",
              "#e8a55a",
              "#c64545",
              "#cc785c",
              "#cc785c",
              "#5db8a6",
              "#5db872",
            ],
          },
        ],
      },
);

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (context: any) =>
          `${context.label}: ${formatTokens(context.parsed)} tokens`,
      },
    },
  },
};
</script>
