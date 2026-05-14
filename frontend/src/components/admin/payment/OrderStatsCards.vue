<template>
  <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
    <!-- Today Revenue -->
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-success/15 p-2">
          <Icon
            name="dollar"
            size="md"
            class="text-success"
            :stroke-width="2"
          />
        </div>
        <div>
          <p class="text-xs font-medium text-muted">
            {{ t("payment.admin.todayRevenue") }}
          </p>
          <p class="text-xl font-bold text-ink">
            ${{ formatMoney(stats.today_amount) }}
          </p>
          <p class="text-xs text-muted">
            {{ stats.today_count }} {{ t("payment.admin.orders") }}
          </p>
        </div>
      </div>
    </div>

    <!-- Total Revenue -->
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-accent-teal/15 p-2">
          <Icon
            name="creditCard"
            size="md"
            class="text-primary-700"
            :stroke-width="2"
          />
        </div>
        <div>
          <p class="text-xs font-medium text-muted">
            {{ t("payment.admin.totalRevenue") }}
          </p>
          <p class="text-xl font-bold text-ink">
            ${{ formatMoney(stats.total_amount) }}
          </p>
          <p class="text-xs text-muted">
            {{ stats.total_count }} {{ t("payment.admin.orders") }}
          </p>
        </div>
      </div>
    </div>

    <!-- Today Orders -->
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-primary-100 p-2">
          <Icon
            name="chart"
            size="md"
            class="text-primary-700"
            :stroke-width="2"
          />
        </div>
        <div>
          <p class="text-xs font-medium text-muted">
            {{ t("payment.admin.todayOrders") }}
          </p>
          <p class="text-xl font-bold text-ink">{{ stats.today_count }}</p>
        </div>
      </div>
    </div>

    <!-- Average Amount -->
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-accent-amber/15 p-2">
          <Icon
            name="chart"
            size="md"
            class="text-warning"
            :stroke-width="2"
          />
        </div>
        <div>
          <p class="text-xs font-medium text-muted">
            {{ t("payment.admin.avgAmount") }}
          </p>
          <p class="text-xl font-bold text-ink">
            ${{ formatMoney(stats.avg_amount) }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";
import type { DashboardStats } from "@/types/payment";

const { t } = useI18n();

defineProps<{
  stats: DashboardStats;
}>();

function formatMoney(value: number): string {
  return value.toFixed(2);
}
</script>
