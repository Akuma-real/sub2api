<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.dataImportTitle')"
    width="normal"
    close-on-click-outside
    @close="handleClose"
  >
    <form
      id="import-data-form"
      class="space-y-4"
      @submit.prevent="handleImport"
    >
      <div class="text-sm text-body">
        {{ t("admin.accounts.dataImportHint") }}
      </div>
      <div class="rounded-lg border border-accent-amber/30 bg-accent-amber/15 p-3 text-xs text-warning">
        {{ t("admin.accounts.dataImportWarning") }}
      </div>

      <div>
        <label class="input-label">{{
          t("admin.accounts.dataImportFile")
        }}</label>
        <div
          data-test="data-import-dropzone"
          :class="[
            'flex items-center justify-between gap-3 rounded-lg border border-dashed px-4 py-3 transition-colors',
            isDragActive
              ? 'border-primary-500 bg-primary-50'
              : 'border-hairline bg-surface-soft',
          ]"
          @dragenter.prevent="handleDragEnter"
          @dragover.prevent="handleDragEnter"
          @dragleave.prevent="handleDragLeave"
          @drop.prevent="handleDrop"
        >
          <div class="flex min-w-0 items-center gap-3">
            <span class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-canvas text-muted">
              <Icon name="upload" size="sm" />
            </span>
            <div class="min-w-0">
              <div class="truncate text-sm text-body">
                {{ fileName || t("admin.accounts.dataImportDropHint") }}
              </div>
              <div class="text-xs text-muted">
                {{
                  fileName
                    ? "JSON (.json)"
                    : t("admin.accounts.dataImportDropSubHint")
                }}
              </div>
            </div>
          </div>
          <button
            type="button"
            class="btn btn-secondary shrink-0"
            @click="openFilePicker"
          >
            <Icon name="upload" size="sm" />
            <span>{{ t("common.chooseFile") }}</span>
          </button>
        </div>
        <input
          ref="fileInput"
          type="file"
          class="hidden"
          accept="application/json,.json"
          @change="handleFileChange"
        />
      </div>

      <div>
        <label class="input-label" for="data-import-default-group">
          {{ t("admin.accounts.dataImportDefaultGroup") }}
        </label>
        <select
          id="data-import-default-group"
          v-model="selectedGroupId"
          class="input-field"
          :disabled="importing || defaultGroupOptions.length === 0"
        >
          <option value="">
            {{ t("admin.accounts.dataImportNoDefaultGroup") }}
          </option>
          <option
            v-for="group in defaultGroupOptions"
            :key="group.id"
            :value="String(group.id)"
          >
            {{ groupOptionLabel(group) }}
          </option>
        </select>
        <p class="mt-1 text-xs text-muted">
          {{ t("admin.accounts.dataImportGroupHint") }}
        </p>
      </div>

      <div
        v-if="result"
        class="space-y-2 rounded-xl border border-hairline p-4"
      >
        <div class="text-sm font-medium text-ink">
          {{ t("admin.accounts.dataImportResult") }}
        </div>
        <div class="text-sm text-body">
          {{ t("admin.accounts.dataImportResultSummary", result) }}
        </div>

        <div v-if="errorItems.length" class="mt-2">
          <div class="text-sm font-medium text-error">
            {{ t("admin.accounts.dataImportErrors") }}
          </div>
          <div
            class="mt-2 max-h-48 overflow-auto rounded-lg bg-surface-soft p-3 font-mono text-xs"
          >
            <div
              v-for="(item, idx) in errorItems"
              :key="idx"
              class="whitespace-pre-wrap"
            >
              {{ item.kind }} {{ item.name || item.proxy_key || "-" }} —
              {{ item.message }}
            </div>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          class="btn btn-secondary"
          type="button"
          :disabled="importing"
          @click="handleClose"
        >
          {{ t("common.cancel") }}
        </button>
        <button
          class="btn btn-primary"
          type="submit"
          form="import-data-form"
          :disabled="importing"
        >
          {{
            importing
              ? t("admin.accounts.dataImporting")
              : t("admin.accounts.dataImportButton")
          }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import BaseDialog from "@/components/common/BaseDialog.vue";
import Icon from "@/components/icons/Icon.vue";
import { adminAPI } from "@/api/admin";
import { useAppStore } from "@/stores/app";
import type { AdminDataImportResult, AdminGroup } from "@/types";

interface Props {
  show: boolean;
  groups?: AdminGroup[];
}

interface Emits {
  (e: "close"): void;
  (e: "imported"): void;
}

const props = withDefaults(defineProps<Props>(), {
  groups: () => [],
});
const emit = defineEmits<Emits>();

const { t } = useI18n();
const appStore = useAppStore();

const importing = ref(false);
const file = ref<File | null>(null);
const isDragActive = ref(false);
const result = ref<AdminDataImportResult | null>(null);
const selectedGroupId = ref("");

const fileInput = ref<HTMLInputElement | null>(null);
const fileName = computed(() => file.value?.name || "");

const errorItems = computed(() => result.value?.errors || []);
const defaultGroupOptions = computed(() =>
  [...props.groups]
    .filter((group) => group.status === "active")
    .sort((a, b) => a.name.localeCompare(b.name)),
);

watch(
  () => props.show,
  (open) => {
    if (open) {
      file.value = null;
      isDragActive.value = false;
      result.value = null;
      selectedGroupId.value = "";
      if (fileInput.value) {
        fileInput.value.value = "";
      }
    }
  },
);

const openFilePicker = () => {
  fileInput.value?.click();
};

const setImportFile = (nextFile: File | null) => {
  file.value = nextFile;
  result.value = null;
};

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  setImportFile(target.files?.[0] || null);
};

const handleDragEnter = () => {
  isDragActive.value = true;
};

const handleDragLeave = () => {
  isDragActive.value = false;
};

const handleDrop = (event: DragEvent) => {
  isDragActive.value = false;
  setImportFile(event.dataTransfer?.files?.[0] || null);
  if (fileInput.value) {
    fileInput.value.value = "";
  }
};

const groupOptionLabel = (group: AdminGroup) => {
  return `${group.name} / ${group.platform}`;
};

const selectedGroupIds = computed(() => {
  const groupId = Number(selectedGroupId.value);
  return Number.isFinite(groupId) && groupId > 0 ? [groupId] : undefined;
});

const handleClose = () => {
  if (importing.value) return;
  emit("close");
};

const readFileAsText = async (sourceFile: File): Promise<string> => {
  if (typeof sourceFile.text === "function") {
    return sourceFile.text();
  }

  if (typeof sourceFile.arrayBuffer === "function") {
    const buffer = await sourceFile.arrayBuffer();
    return new TextDecoder().decode(buffer);
  }

  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result ?? ""));
    reader.onerror = () =>
      reject(reader.error || new Error("Failed to read file"));
    reader.readAsText(sourceFile);
  });
};

const handleImport = async () => {
  if (!file.value) {
    appStore.showError(t("admin.accounts.dataImportSelectFile"));
    return;
  }

  importing.value = true;
  try {
    const text = await readFileAsText(file.value);
    const dataPayload = JSON.parse(text);

    const res = await adminAPI.accounts.importData({
      data: dataPayload,
      group_ids: selectedGroupIds.value,
      skip_default_group_bind: true,
    });

    result.value = res;

    const msgParams: Record<string, unknown> = {
      account_created: res.account_created,
      account_failed: res.account_failed,
      proxy_created: res.proxy_created,
      proxy_reused: res.proxy_reused,
      proxy_failed: res.proxy_failed,
    };
    if (res.account_failed > 0 || res.proxy_failed > 0) {
      appStore.showError(
        t("admin.accounts.dataImportCompletedWithErrors", msgParams),
      );
    } else {
      appStore.showSuccess(t("admin.accounts.dataImportSuccess", msgParams));
      emit("imported");
    }
  } catch (error: any) {
    if (error instanceof SyntaxError) {
      appStore.showError(t("admin.accounts.dataImportParseFailed"));
    } else {
      appStore.showError(
        error?.message || t("admin.accounts.dataImportFailed"),
      );
    }
  } finally {
    importing.value = false;
  }
};
</script>
