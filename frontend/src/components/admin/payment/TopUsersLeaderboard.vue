<template>
  <div class="card p-4">
    <h3 class="mb-4 text-sm font-semibold text-ink">
      {{ t("payment.admin.topUsers") }}
    </h3>
    <div
      v-if="!users?.length"
      class="flex h-32 items-center justify-center text-sm text-muted"
    >
      {{ t("payment.admin.noData") }}
    </div>
    <div v-else class="space-y-2">
      <div
        v-for="(user, idx) in users"
        :key="user.user_id"
        class="flex items-center justify-between rounded-lg px-3 py-2 hover:bg-surface-soft"
      >
        <div class="flex items-center gap-3">
          <span
            :class="[
              'flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold',
              rankClass(idx),
            ]"
          >
            {{ idx + 1 }}
          </span>
          <span class="text-sm text-body">{{ user.email }}</span>
        </div>
        <span class="text-sm font-medium text-ink">
          ${{ user.amount.toFixed(2) }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";

const { t } = useI18n();

defineProps<{
  users: { user_id: number; email: string; amount: number }[];
}>();

function rankClass(idx: number): string {
  if (idx === 0) return "bg-accent-amber/15 text-warning ";
  if (idx === 1) return "bg-hairline text-body ";
  if (idx === 2) return "bg-accent-amber/15 text-warning ";
  return "bg-surface-card text-muted ";
}
</script>
