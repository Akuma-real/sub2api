<template>
  <div class="card p-4">
    <h3 class="mb-4 text-sm font-semibold text-ink">
      {{ t("payment.admin.paymentDistribution") }}
    </h3>
    <div
      v-if="!methods?.length"
      class="flex h-32 items-center justify-center text-sm text-muted"
    >
      {{ t("payment.admin.noData") }}
    </div>
    <div v-else class="space-y-3">
      <div v-for="method in methods" :key="method.type" class="space-y-1">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span
              :class="[
                'inline-block h-3 w-3 rounded-full',
                colorMap[method.type] || 'bg-muted-soft',
              ]"
            ></span>
            <span class="text-sm text-body">
              {{ t("payment.methods." + method.type, method.type) }}
            </span>
          </div>
          <div class="text-right">
            <span class="text-sm font-medium text-ink">
              ${{ method.amount.toFixed(2) }}
            </span>
            <span class="ml-2 text-xs text-muted"> ({{ method.count }}) </span>
          </div>
        </div>
        <div class="h-2 w-full overflow-hidden rounded-full bg-surface-card">
          <div
            :class="[
              'h-full rounded-full transition-all',
              barColorMap[method.type] || 'bg-muted-soft',
            ]"
            :style="{ width: barWidth(method.amount) + '%' }"
          ></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const props = defineProps<{
  methods: { type: string; amount: number; count: number }[];
}>();

const colorMap: Record<string, string> = {
  alipay: "bg-accent-teal",
  wxpay: "bg-success",
  alipay_direct: "bg-accent-teal",
  wxpay_direct: "bg-success",
  stripe: "bg-primary-500",
};

const barColorMap: Record<string, string> = {
  alipay: "bg-accent-teal",
  wxpay: "bg-success",
  alipay_direct: "bg-accent-teal",
  wxpay_direct: "bg-success",
  stripe: "bg-primary-500",
};

const maxAmount = computed(() => {
  if (!props.methods?.length) return 1;
  return Math.max(...props.methods.map((m) => m.amount), 1);
});

function barWidth(amount: number): number {
  return Math.min((amount / maxAmount.value) * 100, 100);
}
</script>
