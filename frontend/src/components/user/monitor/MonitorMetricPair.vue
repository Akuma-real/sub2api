<template>
  <div class="mt-5 grid grid-cols-2 gap-2">
    <div class="rounded-xl p-3 bg-surface-soft/80 border border-hairline-soft">
      <div
        class="flex items-center gap-1.5 text-[10px] font-semibold uppercase tracking-wider text-muted-soft"
      >
        <Icon :name="primaryIcon" size="xs" />
        <span>{{ primaryLabel }}</span>
      </div>
      <div class="mt-1.5 text-lg font-bold font-mono tabular-nums text-ink">
        {{ primaryValue
        }}<span class="text-xs font-normal text-muted-soft ml-0.5">{{
          primaryUnit
        }}</span>
      </div>
    </div>
    <div class="rounded-xl p-3 bg-surface-soft/80 border border-hairline-soft">
      <div
        class="flex items-center gap-1.5 text-[10px] font-semibold uppercase tracking-wider text-muted-soft"
      >
        <Icon :name="secondaryIcon" size="xs" />
        <span>{{ secondaryLabel }}</span>
      </div>
      <div class="mt-1.5 text-lg font-bold font-mono tabular-nums text-ink">
        {{ secondaryValue
        }}<span class="text-xs font-normal text-muted-soft ml-0.5">{{
          secondaryUnit
        }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from "@/components/icons/Icon.vue";

defineProps<{
  primaryLabel: string;
  primaryValue: string;
  primaryUnit: string;
  primaryIcon: "bolt" | "globe" | "clock" | "link";
  secondaryLabel: string;
  secondaryValue: string;
  secondaryUnit: string;
  secondaryIcon: "bolt" | "globe" | "clock" | "link";
}>();
</script>
