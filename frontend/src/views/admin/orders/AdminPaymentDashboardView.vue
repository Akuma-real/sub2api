<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header with Day Switcher -->
      <div class="flex items-center justify-end">
        <div class="flex items-center gap-2">
          <div class="flex rounded-lg border border-hairline">
            <button
              v-for="d in DAYS_OPTIONS"
              :key="d"
              type="button"
              class="px-3 py-1.5 text-xs font-medium transition-colors first:rounded-l-lg last:rounded-r-lg"
              :class="
                days === d
                  ? 'bg-primary-600 text-on-primary'
                  : 'text-body hover:bg-surface-card '
              "
              @click="days = d"
            >
              {{ d }}{{ t("payment.admin.daySuffix") }}
            </button>
          </div>
          <button
            @click="loadDashboard"
            :disabled="loading"
            class="btn btn-secondary"
            :title="t('common.refresh')"
          >
            <Icon
              name="refresh"
              size="md"
              :class="loading ? 'animate-spin' : ''"
            />
          </button>
        </div>
      </div>

      <!-- Dashboard Content -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>
      <template v-else-if="stats">
        <OrderStatsCards :stats="stats" />
        <DailyRevenueChart
          :data="stats.daily_series || []"
          :loading="loading"
        />
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <div class="card p-4">
            <h3 class="mb-4 text-sm font-semibold text-ink">
              {{ t("payment.admin.paymentDistribution") }}
            </h3>
            <div
              v-if="!stats.payment_methods?.length"
              class="flex h-32 items-center justify-center text-sm text-muted"
            >
              {{ t("payment.admin.noData") }}
            </div>
            <div v-else class="space-y-3">
              <div
                v-for="method in stats.payment_methods"
                :key="method.type"
                class="flex items-center justify-between"
              >
                <div class="flex items-center gap-2">
                  <span
                    :class="[
                      'inline-block h-3 w-3 rounded-full',
                      methodColor(method.type),
                    ]"
                  ></span>
                  <span class="text-sm text-body">{{
                    t("payment.methods." + method.type, method.type)
                  }}</span>
                </div>
                <div class="text-right">
                  <span class="text-sm font-medium text-ink"
                    >&yen;{{ method.amount.toFixed(2) }}</span
                  >
                  <span class="ml-2 text-xs text-muted"
                    >({{ method.count }})</span
                  >
                </div>
              </div>
            </div>
          </div>
          <div class="card p-4">
            <h3 class="mb-4 text-sm font-semibold text-ink">
              {{ t("payment.admin.topUsers") }}
            </h3>
            <div
              v-if="!stats.top_users?.length"
              class="flex h-32 items-center justify-center text-sm text-muted"
            >
              {{ t("payment.admin.noData") }}
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="(user, idx) in stats.top_users"
                :key="user.user_id"
                class="flex items-center justify-between rounded-lg px-3 py-2 hover:bg-surface-soft"
              >
                <div class="flex items-center gap-3">
                  <span
                    :class="[
                      'flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold',
                      rankClass(idx),
                    ]"
                    >{{ idx + 1 }}</span
                  >
                  <span class="text-sm text-body">{{ user.email }}</span>
                </div>
                <span class="text-sm font-medium text-ink"
                  >&yen;{{ user.amount.toFixed(2) }}</span
                >
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { useAppStore } from "@/stores/app";
import { adminPaymentAPI } from "@/api/admin/payment";
import { extractI18nErrorMessage } from "@/utils/apiError";
import type { DashboardStats } from "@/types/payment";
import AppLayout from "@/components/layout/AppLayout.vue";
import LoadingSpinner from "@/components/common/LoadingSpinner.vue";
import Icon from "@/components/icons/Icon.vue";
import OrderStatsCards from "@/components/admin/payment/OrderStatsCards.vue";
import DailyRevenueChart from "@/components/admin/payment/DailyRevenueChart.vue";

const { t } = useI18n();
const appStore = useAppStore();

const DAYS_OPTIONS = [7, 30, 90] as const;
const days = ref<number>(30);
const loading = ref(false);
const stats = ref<DashboardStats | null>(null);

function methodColor(type: string): string {
  const c: Record<string, string> = {
    alipay: "bg-accent-teal",
    wxpay: "bg-success",
    alipay_direct: "bg-accent-teal",
    wxpay_direct: "bg-success",
    stripe: "bg-primary-500",
  };
  return c[type] || "bg-muted-soft";
}

function rankClass(idx: number): string {
  if (idx === 0) return "bg-accent-amber/15 text-warning ";
  if (idx === 1) return "bg-hairline text-body ";
  if (idx === 2) return "bg-accent-amber/15 text-warning ";
  return "bg-surface-card text-muted ";
}

async function loadDashboard() {
  loading.value = true;
  try {
    const res = await adminPaymentAPI.getDashboard(days.value);
    stats.value = res.data;
  } catch (err: unknown) {
    appStore.showError(
      extractI18nErrorMessage(err, t, "payment.errors", t("common.error")),
    );
  } finally {
    loading.value = false;
  }
}

watch(days, () => loadDashboard());
onMounted(() => loadDashboard());
</script>
