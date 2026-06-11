<template>
  <div
    class="rounded-lg border border-warning/30 bg-warning/10 p-4 text-left"
    role="alert"
  >
    <div class="flex items-start gap-3">
      <span
        class="mt-0.5 flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-canvas text-warning ring-1 ring-warning/20"
      >
        <Icon name="shield" size="sm" />
      </span>
      <div class="min-w-0 flex-1">
        <p class="text-sm font-semibold text-ink">
          {{ copy.title }}
        </p>
        <p class="mt-1 text-sm leading-6 text-body">
          {{ copy.message }}
        </p>
        <p v-if="detectedFrom" class="mt-2 text-xs text-muted">
          {{ detectedText }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import Icon from "@/components/icons/Icon.vue";
import type { RegionRestrictionCopy } from "@/composables/useRegionRestriction";

const props = defineProps<{
  detectedFrom?: string;
  copy: RegionRestrictionCopy;
}>();

const detectedText = computed(() =>
  props.copy.detectedText.replace("{country}", props.detectedFrom || ""),
);
</script>
