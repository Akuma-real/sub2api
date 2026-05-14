<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import BaseDialog from "@/components/common/BaseDialog.vue";
import Pagination from "@/components/common/Pagination.vue";
import { useClipboard } from "@/composables/useClipboard";
import { useAppStore } from "@/stores";
import {
  opsAPI,
  type OpsRequestDetailsParams,
  type OpsRequestDetail,
} from "@/api/admin/ops";
import { parseTimeRangeMinutes, formatDateTime } from "../utils/opsFormatters";

export interface OpsRequestDetailsPreset {
  title: string;
  kind?: OpsRequestDetailsParams["kind"];
  sort?: OpsRequestDetailsParams["sort"];
  min_duration_ms?: number;
  max_duration_ms?: number;
}

interface Props {
  modelValue: boolean;
  timeRange: string;
  preset: OpsRequestDetailsPreset;
  platform?: string;
  groupId?: number | null;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "openErrorDetail", errorId: number): void;
}>();

const { t } = useI18n();
const appStore = useAppStore();
const { copyToClipboard } = useClipboard();

const loading = ref(false);
const items = ref<OpsRequestDetail[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);

const close = () => emit("update:modelValue", false);

const rangeLabel = computed(() => {
  const minutes = parseTimeRangeMinutes(props.timeRange);
  if (minutes >= 60)
    return t("admin.ops.requestDetails.rangeHours", {
      n: Math.round(minutes / 60),
    });
  return t("admin.ops.requestDetails.rangeMinutes", { n: minutes });
});

function buildTimeParams(): Pick<
  OpsRequestDetailsParams,
  "start_time" | "end_time"
> {
  const minutes = parseTimeRangeMinutes(props.timeRange);
  const endTime = new Date();
  const startTime = new Date(endTime.getTime() - minutes * 60 * 1000);
  return {
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString(),
  };
}

const fetchData = async () => {
  if (!props.modelValue) return;
  loading.value = true;
  try {
    const params: OpsRequestDetailsParams = {
      ...buildTimeParams(),
      page: page.value,
      page_size: pageSize.value,
      kind: props.preset.kind ?? "all",
      sort: props.preset.sort ?? "created_at_desc",
    };

    const platform = (props.platform || "").trim();
    if (platform) params.platform = platform;
    if (typeof props.groupId === "number" && props.groupId > 0)
      params.group_id = props.groupId;

    if (typeof props.preset.min_duration_ms === "number")
      params.min_duration_ms = props.preset.min_duration_ms;
    if (typeof props.preset.max_duration_ms === "number")
      params.max_duration_ms = props.preset.max_duration_ms;

    const res = await opsAPI.listRequestDetails(params);
    items.value = res.items || [];
    total.value = res.total || 0;
  } catch (e: any) {
    console.error(
      "[OpsRequestDetailsModal] Failed to fetch request details",
      e,
    );
    appStore.showError(
      e?.message || t("admin.ops.requestDetails.failedToLoad"),
    );
    items.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
  }
};

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      page.value = 1;
      pageSize.value = 10;
      fetchData();
    }
  },
);

watch(
  () => [
    props.timeRange,
    props.platform,
    props.groupId,
    props.preset.kind,
    props.preset.sort,
    props.preset.min_duration_ms,
    props.preset.max_duration_ms,
  ],
  () => {
    if (!props.modelValue) return;
    page.value = 1;
    fetchData();
  },
);

function handlePageChange(next: number) {
  page.value = next;
  fetchData();
}

function handlePageSizeChange(next: number) {
  pageSize.value = next;
  page.value = 1;
  fetchData();
}

async function handleCopyRequestId(requestId: string) {
  const ok = await copyToClipboard(
    requestId,
    t("admin.ops.requestDetails.requestIdCopied"),
  );
  if (ok) return;
  // `useClipboard` already shows toast on failure; this keeps UX consistent with older ops modal.
  appStore.showWarning(t("admin.ops.requestDetails.copyFailed"));
}

function openErrorDetail(errorId: number | null | undefined) {
  if (!errorId) return;
  close();
  emit("openErrorDetail", errorId);
}

const kindBadgeClass = (kind: string) => {
  if (kind === "error") return "bg-error/15 text-error ";
  return "bg-success/15 text-success ";
};
</script>

<template>
  <BaseDialog
    :show="modelValue"
    :title="props.preset.title || t('admin.ops.requestDetails.title')"
    width="full"
    @close="close"
  >
    <template #default>
      <div class="flex h-full min-h-0 flex-col">
        <div class="mb-4 flex flex-shrink-0 items-center justify-between">
          <div class="text-xs text-muted">
            {{
              t("admin.ops.requestDetails.rangeLabel", { range: rangeLabel })
            }}
          </div>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            @click="fetchData"
          >
            {{ t("common.refresh") }}
          </button>
        </div>

        <!-- Loading -->
        <div
          v-if="loading"
          class="flex flex-1 items-center justify-center py-16"
        >
          <div class="flex flex-col items-center gap-3">
            <svg
              class="h-8 w-8 animate-spin text-accent-teal"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            <span class="text-sm font-medium text-muted">{{
              t("common.loading")
            }}</span>
          </div>
        </div>

        <!-- Table -->
        <div v-else class="flex min-h-0 flex-1 flex-col">
          <div
            v-if="items.length === 0"
            class="rounded-xl border border-dashed border-hairline p-10 text-center"
          >
            <div class="text-sm font-medium text-body">
              {{ t("admin.ops.requestDetails.empty") }}
            </div>
            <div class="mt-1 text-xs text-muted-soft">
              {{ t("admin.ops.requestDetails.emptyHint") }}
            </div>
          </div>

          <div
            v-else
            class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-hairline"
          >
            <div class="min-h-0 flex-1 overflow-auto">
              <table class="min-w-full divide-y divide-hairline">
                <thead class="sticky top-0 z-10 bg-surface-soft">
                  <tr>
                    <th
                      class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
                    >
                      {{ t("admin.ops.requestDetails.table.time") }}
                    </th>
                    <th
                      class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
                    >
                      {{ t("admin.ops.requestDetails.table.kind") }}
                    </th>
                    <th
                      class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
                    >
                      {{ t("admin.ops.requestDetails.table.platform") }}
                    </th>
                    <th
                      class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
                    >
                      {{ t("admin.ops.requestDetails.table.model") }}
                    </th>
                    <th
                      class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
                    >
                      {{ t("admin.ops.requestDetails.table.duration") }}
                    </th>
                    <th
                      class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
                    >
                      {{ t("admin.ops.requestDetails.table.status") }}
                    </th>
                    <th
                      class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
                    >
                      {{ t("admin.ops.requestDetails.table.requestId") }}
                    </th>
                    <th
                      class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-muted"
                    >
                      {{ t("admin.ops.requestDetails.table.actions") }}
                    </th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-hairline bg-canvas">
                  <tr
                    v-for="(row, idx) in items"
                    :key="idx"
                    class="hover:bg-surface-soft"
                  >
                    <td class="whitespace-nowrap px-4 py-3 text-xs text-body">
                      {{ formatDateTime(row.created_at) }}
                    </td>
                    <td class="whitespace-nowrap px-4 py-3">
                      <span
                        class="rounded-full px-2 py-1 text-[10px] font-bold"
                        :class="kindBadgeClass(row.kind)"
                      >
                        {{
                          row.kind === "error"
                            ? t("admin.ops.requestDetails.kind.error")
                            : t("admin.ops.requestDetails.kind.success")
                        }}
                      </span>
                    </td>
                    <td
                      class="whitespace-nowrap px-4 py-3 text-xs font-medium text-body"
                    >
                      {{ (row.platform || "unknown").toUpperCase() }}
                    </td>
                    <td
                      class="max-w-[240px] truncate px-4 py-3 text-xs text-body"
                      :title="row.model || ''"
                    >
                      {{ row.model || "-" }}
                    </td>
                    <td class="whitespace-nowrap px-4 py-3 text-xs text-body">
                      {{
                        typeof row.duration_ms === "number"
                          ? `${row.duration_ms} ms`
                          : "-"
                      }}
                    </td>
                    <td class="whitespace-nowrap px-4 py-3 text-xs text-body">
                      {{ row.status_code ?? "-" }}
                    </td>
                    <td class="px-4 py-3">
                      <div
                        v-if="row.request_id"
                        class="flex items-center gap-2"
                      >
                        <span
                          class="max-w-[220px] truncate font-mono text-[11px] text-body"
                          :title="row.request_id"
                        >
                          {{ row.request_id }}
                        </span>
                        <button
                          class="rounded-md bg-surface-card px-2 py-1 text-[10px] font-bold text-body hover:bg-hairline"
                          @click="handleCopyRequestId(row.request_id)"
                        >
                          {{ t("admin.ops.requestDetails.copy") }}
                        </button>
                      </div>
                      <span v-else class="text-xs text-muted-soft">-</span>
                    </td>
                    <td class="whitespace-nowrap px-4 py-3 text-right">
                      <button
                        v-if="row.kind === 'error' && row.error_id"
                        class="rounded-lg bg-error/15 px-3 py-1.5 text-xs font-bold text-error hover:bg-error/15"
                        @click="openErrorDetail(row.error_id)"
                      >
                        {{ t("admin.ops.requestDetails.viewError") }}
                      </button>
                      <span v-else class="text-xs text-muted-soft">-</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <Pagination
              :total="total"
              :page="page"
              :page-size="pageSize"
              @update:page="handlePageChange"
              @update:pageSize="handlePageSizeChange"
            />
          </div>
        </div>
      </div>
    </template>
  </BaseDialog>
</template>
