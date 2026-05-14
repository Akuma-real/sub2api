<template>
  <div class="flex min-w-0 flex-1 items-start justify-between gap-3">
    <!-- Left: name + description -->
    <div
      class="flex min-w-0 flex-1 flex-col items-start"
      :title="description || undefined"
    >
      <!-- Row 1: platform badge (name bold) -->
      <GroupBadge
        :name="name"
        :platform="platform"
        :subscription-type="subscriptionType"
        :show-rate="false"
        class="groupOptionItemBadge"
      />
      <!-- Row 2: description with top spacing -->
      <span
        v-if="description"
        class="mt-1.5 w-full text-left text-xs leading-relaxed text-muted line-clamp-2"
      >
        {{ description }}
      </span>
    </div>

    <!-- Right: rate pill + checkmark (vertically centered to first row) -->
    <div class="flex shrink-0 items-center gap-2 pt-0.5">
      <!-- Rate pill (platform color) -->
      <span
        v-if="rateMultiplier !== undefined"
        :class="[
          'inline-flex items-center whitespace-nowrap rounded-full px-3 py-1 text-xs font-semibold',
          ratePillClass,
        ]"
      >
        <template v-if="hasCustomRate">
          <span class="mr-1 line-through opacity-50"
            >{{ rateMultiplier }}x</span
          >
          <span class="font-bold">{{ userRateMultiplier }}x</span>
        </template>
        <template v-else> {{ rateMultiplier }}x 倍率 </template>
      </span>
      <!-- Checkmark -->
      <svg
        v-if="showCheckmark && selected"
        class="h-4 w-4 shrink-0 text-primary-600"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        stroke-width="2"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M5 13l4 4L19 7"
        />
      </svg>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import GroupBadge from "./GroupBadge.vue";
import type { SubscriptionType, GroupPlatform } from "@/types";

interface Props {
  name: string;
  platform: GroupPlatform;
  subscriptionType?: SubscriptionType;
  rateMultiplier?: number;
  userRateMultiplier?: number | null;
  description?: string | null;
  selected?: boolean;
  showCheckmark?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  subscriptionType: "standard",
  selected: false,
  showCheckmark: true,
  userRateMultiplier: null,
});

// Whether user has a custom rate different from default
const hasCustomRate = computed(() => {
  return (
    props.userRateMultiplier !== null &&
    props.userRateMultiplier !== undefined &&
    props.rateMultiplier !== undefined &&
    props.userRateMultiplier !== props.rateMultiplier
  );
});

// Rate pill color matches platform badge color
const ratePillClass = computed(() => {
  switch (props.platform) {
    case "anthropic":
      return "bg-accent-amber/15 text-warning ";
    case "openai":
      return "bg-success/15 text-success ";
    case "gemini":
      return "bg-accent-teal/15 text-accent-teal ";
    default: // antigravity and others
      return "bg-primary-100 text-primary-700 ";
  }
});
</script>

<style scoped>
/* Bold the group name inside GroupBadge when used in dropdown option */
.groupOptionItemBadge :deep(span.truncate) {
  font-weight: 600;
}
</style>
