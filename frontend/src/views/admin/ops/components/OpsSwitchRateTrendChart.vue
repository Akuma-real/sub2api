<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import {
  Chart as ChartJS,
  CategoryScale,
  Filler,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  Title,
  Tooltip,
} from "chart.js";
import { Line } from "vue-chartjs";
import type { OpsThroughputTrendPoint } from "@/api/admin/ops";
import type { ChartState } from "../types";
import { formatHistoryLabel, sumNumbers } from "../utils/opsFormatters";
import HelpTooltip from "@/components/common/HelpTooltip.vue";
import EmptyState from "@/components/common/EmptyState.vue";

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  CategoryScale,
  Filler,
);

interface Props {
  points: OpsThroughputTrendPoint[];
  loading: boolean;
  timeRange: string;
  fullscreen?: boolean;
}

const props = defineProps<Props>();
const { t } = useI18n();

const colors = {
  teal: "#5db8a6",
  tealAlpha: "#5db8a620",
  grid: "#f5f0e8",
  text: "#6c6a64",
};

const totalRequests = computed(() =>
  sumNumbers(props.points.map((p) => p.request_count)),
);

const chartData = computed(() => {
  if (!props.points.length || totalRequests.value <= 0) return null;
  return {
    labels: props.points.map((p) =>
      formatHistoryLabel(p.bucket_start, props.timeRange),
    ),
    datasets: [
      {
        label: t("admin.ops.switchRate"),
        data: props.points.map((p) => {
          const requests = p.request_count ?? 0;
          const switches = p.switch_count ?? 0;
          if (requests <= 0) return 0;
          return switches / requests;
        }),
        borderColor: colors.teal,
        backgroundColor: colors.tealAlpha,
        fill: true,
        tension: 0.35,
        pointRadius: 0,
        pointHitRadius: 10,
      },
    ],
  };
});

const state = computed<ChartState>(() => {
  if (chartData.value) return "ready";
  if (props.loading) return "loading";
  return "empty";
});

const options = computed(() => {
  const c = colors;
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { intersect: false, mode: "index" as const },
    plugins: {
      legend: {
        position: "top" as const,
        align: "end" as const,
        labels: {
          color: c.text,
          usePointStyle: true,
          boxWidth: 6,
          font: { size: 10 },
        },
      },
      tooltip: {
        backgroundColor: "#faf9f5",
        titleColor: "#141413",
        bodyColor: "#3d3d3a",
        borderColor: c.grid,
        borderWidth: 1,
        padding: 10,
        displayColors: true,
        callbacks: {
          label: (context: any) => {
            const value =
              typeof context?.parsed?.y === "number" ? context.parsed.y : 0;
            return `${t("admin.ops.switchRate")}: ${value.toFixed(3)}`;
          },
        },
      },
    },
    scales: {
      x: {
        type: "category" as const,
        grid: { display: false },
        ticks: {
          color: c.text,
          font: { size: 10 },
          maxTicksLimit: 8,
          autoSkip: true,
          autoSkipPadding: 10,
        },
      },
      y: {
        type: "linear" as const,
        display: true,
        position: "left" as const,
        grid: { color: c.grid, borderDash: [4, 4] },
        ticks: {
          color: c.text,
          font: { size: 10 },
          callback: (value: any) => Number(value).toFixed(3),
        },
      },
    },
  };
});
</script>

<template>
  <div
    class="flex h-full flex-col rounded-lg bg-canvas p-6 shadow-sm ring-1 ring-ink/5"
  >
    <div class="mb-4 flex shrink-0 items-center justify-between">
      <h3 class="flex items-center gap-2 text-sm font-bold text-ink">
        <svg
          class="h-4 w-4 text-accent-teal"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M7 7h10M7 12h6m-6 5h3"
          />
        </svg>
        {{ t("admin.ops.switchRateTrend") }}
        <HelpTooltip
          v-if="!props.fullscreen"
          :content="t('admin.ops.tooltips.switchRateTrend')"
        />
      </h3>
    </div>

    <div class="min-h-0 flex-1">
      <Line
        v-if="state === 'ready' && chartData"
        :data="chartData"
        :options="options"
      />
      <div v-else class="flex h-full items-center justify-center">
        <div
          v-if="state === 'loading'"
          class="animate-pulse text-sm text-muted-soft"
        >
          {{ t("common.loading") }}
        </div>
        <EmptyState
          v-else
          :title="t('common.noData')"
          :description="t('admin.ops.charts.emptyRequest')"
        />
      </div>
    </div>
  </div>
</template>
