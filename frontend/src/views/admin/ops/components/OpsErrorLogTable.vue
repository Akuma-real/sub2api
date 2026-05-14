<template>
  <div class="flex h-full min-h-0 flex-col bg-canvas">
    <!-- Loading State -->
    <div v-if="loading" class="flex flex-1 items-center justify-center py-10">
      <div
        class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"
      ></div>
    </div>

    <!-- Table Container -->
    <div v-else class="flex min-h-0 flex-1 flex-col">
      <div class="min-h-0 flex-1 overflow-auto border-b border-hairline">
        <table class="w-full border-separate border-spacing-0">
          <thead class="sticky top-0 z-10 bg-surface-soft">
            <tr>
              <th
                class="border-b border-hairline px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
              >
                {{ t("admin.ops.errorLog.time") }}
              </th>
              <th
                class="border-b border-hairline px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
              >
                {{ t("admin.ops.errorLog.type") }}
              </th>
              <th
                class="border-b border-hairline px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
              >
                {{ t("admin.ops.errorLog.endpoint") }}
              </th>
              <th
                class="border-b border-hairline px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
              >
                {{ t("admin.ops.errorLog.platform") }}
              </th>
              <th
                class="border-b border-hairline px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
              >
                {{ t("admin.ops.errorLog.model") }}
              </th>
              <th
                class="border-b border-hairline px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
              >
                {{ t("admin.ops.errorLog.group") }}
              </th>
              <th
                class="border-b border-hairline px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
              >
                {{ t("admin.ops.errorLog.user") }}
              </th>
              <th
                class="border-b border-hairline px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
              >
                {{ t("admin.ops.errorLog.status") }}
              </th>
              <th
                class="border-b border-hairline px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted"
              >
                {{ t("admin.ops.errorLog.message") }}
              </th>
              <th
                class="border-b border-hairline px-4 py-2.5 text-right text-[11px] font-bold uppercase tracking-wider text-muted"
              >
                {{ t("admin.ops.errorLog.action") }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-hairline-soft">
            <tr v-if="rows.length === 0">
              <td
                colspan="10"
                class="py-12 text-center text-sm text-muted-soft"
              >
                {{ t("admin.ops.errorLog.noErrors") }}
              </td>
            </tr>

            <tr
              v-for="log in rows"
              :key="log.id"
              class="group cursor-pointer transition-colors hover:bg-surface-soft/80"
              @click="emit('openErrorDetail', log.id)"
            >
              <!-- Time -->
              <td class="whitespace-nowrap px-4 py-2">
                <el-tooltip
                  :content="log.request_id || log.client_request_id"
                  placement="top"
                  :show-after="500"
                >
                  <span class="font-mono text-xs font-medium text-ink">
                    {{ formatDateTime(log.created_at).split(" ")[1] }}
                  </span>
                </el-tooltip>
              </td>

              <!-- Type -->
              <td class="whitespace-nowrap px-4 py-2">
                <span
                  :class="[
                    'inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-bold ring-1 ring-inset',
                    getTypeBadge(log).className,
                  ]"
                >
                  {{ getTypeBadge(log).label }}
                </span>
              </td>

              <!-- Endpoint -->
              <td class="px-4 py-2">
                <div class="max-w-[160px]">
                  <el-tooltip
                    v-if="log.inbound_endpoint"
                    :content="formatEndpointTooltip(log)"
                    placement="top"
                    :show-after="500"
                  >
                    <span class="truncate font-mono text-[11px] text-body">
                      {{ log.inbound_endpoint }}
                    </span>
                  </el-tooltip>
                  <span v-else class="text-xs text-muted-soft">-</span>
                </div>
              </td>

              <!-- Platform -->
              <td class="whitespace-nowrap px-4 py-2">
                <span
                  class="inline-flex items-center rounded bg-surface-card px-1.5 py-0.5 text-[10px] font-bold uppercase text-body"
                >
                  {{ log.platform || "-" }}
                </span>
              </td>

              <!-- Model -->
              <td class="px-4 py-2">
                <div class="max-w-[160px]">
                  <template v-if="hasModelMapping(log)">
                    <el-tooltip
                      :content="modelMappingTooltip(log)"
                      placement="top"
                      :show-after="500"
                    >
                      <span
                        class="flex items-center gap-1 truncate font-mono text-[11px] text-body"
                      >
                        <span class="truncate">{{ log.requested_model }}</span>
                        <span class="flex-shrink-0 text-muted-soft">→</span>
                        <span class="truncate text-primary-600">{{
                          log.upstream_model
                        }}</span>
                      </span>
                    </el-tooltip>
                  </template>
                  <template v-else>
                    <span
                      v-if="displayModel(log)"
                      class="truncate font-mono text-[11px] text-body"
                      :title="displayModel(log)"
                    >
                      {{ displayModel(log) }}
                    </span>
                    <span v-else class="text-xs text-muted-soft">-</span>
                  </template>
                </div>
              </td>

              <!-- Group -->
              <td class="px-4 py-2">
                <el-tooltip
                  v-if="log.group_id"
                  :content="t('admin.ops.errorLog.id') + ' ' + log.group_id"
                  placement="top"
                  :show-after="500"
                >
                  <span
                    class="max-w-[100px] truncate text-xs font-medium text-ink"
                  >
                    {{ log.group_name || "-" }}
                  </span>
                </el-tooltip>
                <span v-else class="text-xs text-muted-soft">-</span>
              </td>

              <!-- User / Account -->
              <td class="px-4 py-2">
                <template v-if="isUpstreamRow(log)">
                  <el-tooltip
                    v-if="log.account_id"
                    :content="
                      t('admin.ops.errorLog.accountId') + ' ' + log.account_id
                    "
                    placement="top"
                    :show-after="500"
                  >
                    <span
                      class="max-w-[100px] truncate text-xs font-medium text-ink"
                    >
                      {{ log.account_name || "-" }}
                    </span>
                  </el-tooltip>
                  <span v-else class="text-xs text-muted-soft">-</span>
                </template>
                <template v-else>
                  <el-tooltip
                    v-if="log.user_id"
                    :content="
                      t('admin.ops.errorLog.userId') + ' ' + log.user_id
                    "
                    placement="top"
                    :show-after="500"
                  >
                    <span
                      class="max-w-[100px] truncate text-xs font-medium text-ink"
                    >
                      {{ log.user_email || "-" }}
                    </span>
                  </el-tooltip>
                  <span v-else class="text-xs text-muted-soft">-</span>
                </template>
              </td>

              <!-- Status -->
              <td class="whitespace-nowrap px-4 py-2">
                <div class="flex items-center gap-1.5">
                  <span
                    :class="[
                      'inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-bold ring-1 ring-inset',
                      getStatusClass(log.status_code),
                    ]"
                  >
                    {{ log.status_code }}
                  </span>
                  <span
                    v-if="log.severity"
                    :class="[
                      'rounded px-1.5 py-0.5 text-[10px] font-bold',
                      getSeverityClass(log.severity),
                    ]"
                  >
                    {{ log.severity }}
                  </span>
                  <span
                    v-if="log.request_type != null && log.request_type > 0"
                    class="rounded bg-surface-card px-1.5 py-0.5 text-[10px] font-bold text-body"
                  >
                    {{ formatRequestType(log.request_type) }}
                  </span>
                </div>
              </td>

              <!-- Message (Response Content) -->
              <td class="px-4 py-2">
                <div class="max-w-[200px]">
                  <p
                    class="truncate text-[11px] font-medium text-body"
                    :title="log.message"
                  >
                    {{ formatSmartMessage(log.message) || "-" }}
                  </p>
                </div>
              </td>

              <!-- Actions -->
              <td class="whitespace-nowrap px-4 py-2 text-right" @click.stop>
                <div class="flex items-center justify-end gap-3">
                  <button
                    type="button"
                    class="text-primary-600 hover:text-primary-700 text-xs font-bold"
                    @click="emit('openErrorDetail', log.id)"
                  >
                    {{ t("admin.ops.errorLog.details") }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div class="bg-surface-soft/50">
        <Pagination
          v-if="total > 0"
          :total="total"
          :page="page"
          :page-size="pageSize"
          @update:page="emit('update:page', $event)"
          @update:pageSize="emit('update:pageSize', $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import Pagination from "@/components/common/Pagination.vue";
import type { OpsErrorLog } from "@/api/admin/ops";
import { getSeverityClass, formatDateTime } from "../utils/opsFormatters";

const { t } = useI18n();

function isUpstreamRow(log: OpsErrorLog): boolean {
  const phase = String(log.phase || "").toLowerCase();
  const owner = String(log.error_owner || "").toLowerCase();
  return phase === "upstream" && owner === "provider";
}

function formatEndpointTooltip(log: OpsErrorLog): string {
  const parts: string[] = [];
  if (log.inbound_endpoint) parts.push(`Inbound: ${log.inbound_endpoint}`);
  if (log.upstream_endpoint) parts.push(`Upstream: ${log.upstream_endpoint}`);
  return parts.join("\n") || "";
}

function hasModelMapping(log: OpsErrorLog): boolean {
  const requested = String(log.requested_model || "").trim();
  const upstream = String(log.upstream_model || "").trim();
  return !!requested && !!upstream && requested !== upstream;
}

function modelMappingTooltip(log: OpsErrorLog): string {
  const requested = String(log.requested_model || "").trim();
  const upstream = String(log.upstream_model || "").trim();
  if (!requested && !upstream) return "";
  if (requested && upstream) return `${requested} → ${upstream}`;
  return upstream || requested;
}

function displayModel(log: OpsErrorLog): string {
  const upstream = String(log.upstream_model || "").trim();
  if (upstream) return upstream;
  const requested = String(log.requested_model || "").trim();
  if (requested) return requested;
  return String(log.model || "").trim();
}

function formatRequestType(type: number | null | undefined): string {
  switch (type) {
    case 1:
      return t("admin.ops.errorLog.requestTypeSync");
    case 2:
      return t("admin.ops.errorLog.requestTypeStream");
    case 3:
      return t("admin.ops.errorLog.requestTypeWs");
    default:
      return "";
  }
}

function getTypeBadge(log: OpsErrorLog): { label: string; className: string } {
  const phase = String(log.phase || "").toLowerCase();
  const owner = String(log.error_owner || "").toLowerCase();

  if (isUpstreamRow(log)) {
    return {
      label: t("admin.ops.errorLog.typeUpstream"),
      className: "bg-error/15 text-error ring-error/20 ",
    };
  }
  if (phase === "request" && owner === "client") {
    return {
      label: t("admin.ops.errorLog.typeRequest"),
      className: "bg-accent-amber/15 text-warning ring-accent-amber/20 ",
    };
  }
  if (phase === "auth" && owner === "client") {
    return {
      label: t("admin.ops.errorLog.typeAuth"),
      className: "bg-accent-teal/10 text-primary-700 ring-primary-500/20 ",
    };
  }
  if (phase === "routing" && owner === "platform") {
    return {
      label: t("admin.ops.errorLog.typeRouting"),
      className: "bg-primary-50 text-primary-700 ring-primary-500/20 ",
    };
  }
  if (phase === "internal" && owner === "platform") {
    return {
      label: t("admin.ops.errorLog.typeInternal"),
      className: "bg-surface-card text-body-strong ring-hairline/20 ",
    };
  }

  const fallback = phase || owner || t("common.unknown");
  return {
    label: fallback,
    className: "bg-surface-soft text-body ring-hairline/10 ",
  };
}

interface Props {
  rows: OpsErrorLog[];
  total: number;
  loading: boolean;
  page: number;
  pageSize: number;
}

interface Emits {
  (e: "openErrorDetail", id: number): void;
  (e: "update:page", value: number): void;
  (e: "update:pageSize", value: number): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();

function getStatusClass(code: number): string {
  if (code >= 500) return "bg-error/15 text-error ring-error/20 ";
  if (code === 429) return "bg-primary-50 text-primary-700 ring-primary-500/20 ";
  if (code >= 400) return "bg-accent-amber/15 text-warning ring-accent-amber/20 ";
  return "bg-surface-soft text-body ring-hairline/20 ";
}

function formatSmartMessage(msg: string): string {
  if (!msg) return "";

  if (msg.startsWith("{") || msg.startsWith("[")) {
    try {
      const obj = JSON.parse(msg);
      if (obj?.error?.message) return String(obj.error.message);
      if (obj?.message) return String(obj.message);
      if (obj?.detail) return String(obj.detail);
      if (typeof obj === "object") return JSON.stringify(obj).substring(0, 150);
    } catch {
      // ignore parse error
    }
  }

  if (msg.includes("context deadline exceeded"))
    return t("admin.ops.errorLog.commonErrors.contextDeadlineExceeded");
  if (msg.includes("connection refused"))
    return t("admin.ops.errorLog.commonErrors.connectionRefused");
  if (msg.toLowerCase().includes("rate limit"))
    return t("admin.ops.errorLog.commonErrors.rateLimit");

  return msg.length > 200 ? msg.substring(0, 200) + "..." : msg;
}
</script>
