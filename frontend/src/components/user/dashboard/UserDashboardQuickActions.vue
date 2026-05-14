<template>
  <div class="card">
    <div class="border-b border-hairline-soft px-6 py-4">
      <h2 class="text-lg font-semibold text-ink">
        {{ t("dashboard.quickActions") }}
      </h2>
    </div>
    <div class="space-y-3 p-4">
      <button
        @click="router.push('/keys')"
        class="group flex w-full items-center gap-4 rounded-xl bg-surface-soft p-4 text-left transition-all duration-200 hover:bg-surface-card"
      >
        <div
          class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-primary-100 transition-transform group-hover:scale-105"
        >
          <Icon name="key" size="lg" class="text-primary-600" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-ink">
            {{ t("dashboard.createApiKey") }}
          </p>
          <p class="text-xs text-muted">{{ t("dashboard.generateNewKey") }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-muted-soft transition-colors group-hover:text-primary-500"
        />
      </button>

      <button
        @click="router.push('/usage')"
        class="group flex w-full items-center gap-4 rounded-xl bg-surface-soft p-4 text-left transition-all duration-200 hover:bg-surface-card"
      >
        <div
          class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-success/15 transition-transform group-hover:scale-105"
        >
          <Icon name="chart" size="lg" class="text-success" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-ink">
            {{ t("dashboard.viewUsage") }}
          </p>
          <p class="text-xs text-muted">
            {{ t("dashboard.checkDetailedLogs") }}
          </p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-muted-soft transition-colors group-hover:text-success"
        />
      </button>

      <button
        @click="router.push('/redeem')"
        class="group flex w-full items-center gap-4 rounded-xl bg-surface-soft p-4 text-left transition-all duration-200 hover:bg-surface-card"
      >
        <div
          class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-accent-amber/15 transition-transform group-hover:scale-105"
        >
          <Icon name="gift" size="lg" class="text-warning" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-ink">
            {{ t("dashboard.redeemCode") }}
          </p>
          <p class="text-xs text-muted">
            {{ t("dashboard.addBalanceWithCode") }}
          </p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-muted-soft transition-colors group-hover:text-warning"
        />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";
const router = useRouter();
const { t } = useI18n();
</script>
