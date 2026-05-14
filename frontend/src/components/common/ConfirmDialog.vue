<template>
  <BaseDialog :show="show" :title="title" width="narrow" @close="handleCancel">
    <div class="space-y-4">
      <p class="text-sm text-body">{{ message }}</p>
      <slot></slot>
    </div>

    <template #footer>
      <div class="flex justify-end space-x-3">
        <button
          @click="handleCancel"
          type="button"
          class="rounded-md border border-hairline bg-canvas px-4 py-2 text-sm font-medium text-body hover:bg-surface-soft focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
        >
          {{ cancelText }}
        </button>
        <button
          @click="handleConfirm"
          type="button"
          :class="[
            'rounded-md px-4 py-2 text-sm font-medium text-on-primary focus:outline-none focus:ring-2 focus:ring-offset-2 ',
            danger
              ? 'bg-error hover:bg-error focus:ring-error'
              : 'bg-primary-600 hover:bg-primary-700 focus:ring-primary-500',
          ]"
        >
          {{ confirmText }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import BaseDialog from "./BaseDialog.vue";

const { t } = useI18n();

interface Props {
  show: boolean;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  danger?: boolean;
}

interface Emits {
  (e: "confirm"): void;
  (e: "cancel"): void;
}

const props = withDefaults(defineProps<Props>(), {
  danger: false,
});

const confirmText = computed(() => props.confirmText || t("common.confirm"));
const cancelText = computed(() => props.cancelText || t("common.cancel"));

const emit = defineEmits<Emits>();

const handleConfirm = () => {
  emit("confirm");
};

const handleCancel = () => {
  emit("cancel");
};
</script>
