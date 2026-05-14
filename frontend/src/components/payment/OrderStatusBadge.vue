<template>
  <span
    class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
    :class="statusClass"
  >
    {{ statusLabel }}
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import type { OrderStatus } from "@/types/payment";

const props = defineProps<{
  status: OrderStatus;
}>();

const { t } = useI18n();

const statusMap: Record<OrderStatus, { key: string; class: string }> = {
  PENDING: {
    key: "payment.status.pending",
    class: "bg-accent-amber/15 text-warning ",
  },
  PAID: {
    key: "payment.status.paid",
    class: "bg-accent-teal/15 text-body-strong ",
  },
  RECHARGING: {
    key: "payment.status.recharging",
    class: "bg-accent-teal/15 text-body-strong ",
  },
  COMPLETED: {
    key: "payment.status.completed",
    class: "bg-success/15 text-success ",
  },
  EXPIRED: {
    key: "payment.status.expired",
    class: "bg-surface-card text-body-strong ",
  },
  CANCELLED: {
    key: "payment.status.cancelled",
    class: "bg-surface-card text-body-strong ",
  },
  FAILED: { key: "payment.status.failed", class: "bg-error/15 text-error " },
  REFUND_REQUESTED: {
    key: "payment.status.refund_requested",
    class: "bg-accent-amber/15 text-warning ",
  },
  REFUNDING: {
    key: "payment.status.refunding",
    class: "bg-accent-amber/15 text-warning ",
  },
  REFUNDED: {
    key: "payment.status.refunded",
    class: "bg-primary-100 text-primary-800 ",
  },
  PARTIALLY_REFUNDED: {
    key: "payment.status.partially_refunded",
    class: "bg-primary-100 text-primary-800 ",
  },
  REFUND_FAILED: {
    key: "payment.status.refund_failed",
    class: "bg-error/15 text-error ",
  },
};

const statusLabel = computed(() => {
  const entry = statusMap[props.status];
  return entry ? t(entry.key) : props.status;
});

const statusClass = computed(() => {
  const entry = statusMap[props.status];
  return entry?.class ?? "bg-surface-card text-body-strong ";
});
</script>
