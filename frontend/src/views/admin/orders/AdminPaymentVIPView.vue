<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-lg font-semibold text-ink">{{ t("payment.admin.vipLevels") }}</h2>
          <div class="flex items-center gap-2">
            <button class="btn btn-secondary" :disabled="levelsLoading" @click="loadLevels">
              <Icon name="refresh" size="md" :class="levelsLoading ? 'animate-spin' : ''" />
            </button>
            <button class="btn btn-primary" @click="openLevelDialog(null)">
              {{ t("payment.admin.createVIPLevel") }}
            </button>
          </div>
        </div>

        <DataTable :columns="levelColumns" :data="levels" :loading="levelsLoading">
          <template #cell-name="{ value }">
            <span class="text-sm font-medium text-primary-700">{{ value }}</span>
          </template>
          <template #cell-price="{ value, row }">
            <span class="font-medium text-ink">¥{{ Number(value || 0).toFixed(2) }}</span>
            <span v-if="row.original_price" class="ml-1 text-xs text-muted line-through">
              ¥{{ Number(row.original_price).toFixed(2) }}
            </span>
          </template>
          <template #cell-discount_multiplier="{ value }">
            <span class="badge badge-success">{{ t("payment.vip.discountValue", { percent: Math.round(Number(value || 1) * 100) }) }}</span>
          </template>
          <template #cell-for_sale="{ value, row }">
            <button
              type="button"
              :class="[
                'relative inline-flex h-5 w-9 rounded-full border-2 border-transparent transition-colors',
                value ? 'bg-primary-500' : 'bg-surface-cream-strong',
              ]"
              @click="toggleForSale(row)"
            >
              <span :class="['inline-block h-4 w-4 rounded-full bg-canvas shadow transition', value ? 'translate-x-4' : 'translate-x-0']" />
            </button>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center gap-2">
              <button class="rounded-lg p-1.5 text-muted hover:bg-accent-teal/10 hover:text-primary-700" @click="openLevelDialog(row)">
                <Icon name="edit" size="sm" />
              </button>
              <button class="rounded-lg p-1.5 text-muted hover:bg-error/15 hover:text-error" @click="confirmDelete(row)">
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </template>
        </DataTable>
      </section>

      <section class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-lg font-semibold text-ink">{{ t("payment.admin.vipUsers") }}</h2>
          <div class="flex items-center gap-2">
            <button class="btn btn-secondary" :disabled="usersLoading" @click="loadUsers">
              <Icon name="refresh" size="md" :class="usersLoading ? 'animate-spin' : ''" />
            </button>
            <button class="btn btn-primary" @click="openAssignDialog">
              <Icon name="plus" size="md" class="mr-2" />
              {{ t("payment.admin.assignVIP") }}
            </button>
          </div>
        </div>
        <DataTable :columns="userColumns" :data="users" :loading="usersLoading">
          <template #cell-user="{ row }">
            <div class="text-sm">
              <div class="font-medium text-ink">{{ row.username || row.email }}</div>
              <div class="text-xs text-muted">{{ row.email }}</div>
            </div>
          </template>
          <template #cell-current="{ row }">
            <div v-if="row.current" class="text-sm">
              <span class="font-medium text-primary-700">
                {{ row.current.level ? vipLevelDisplayName(row.current.level) : `VIP #${row.current.vip_level_id}` }}
              </span>
              <div class="text-xs text-muted">{{ formatDate(row.current.expires_at) }}</div>
            </div>
            <span v-else class="text-sm text-muted">{{ t("payment.vip.none") }}</span>
          </template>
          <template #cell-total_savings_usd="{ value }">
            <span class="font-medium text-success">${{ Number(value || 0).toFixed(4) }}</span>
          </template>
        </DataTable>
      </section>
    </div>

    <BaseDialog :show="showLevelDialog" :title="editingLevel ? t('payment.admin.editVIPLevel') : t('payment.admin.createVIPLevel')" width="wide" @close="showLevelDialog = false">
      <form id="vip-level-form" class="space-y-4" @submit.prevent="saveLevel">
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t("payment.admin.vipLevelName") }}</label>
            <input v-model="levelForm.name" class="input" required />
          </div>
          <div>
            <label class="input-label">{{ t("payment.admin.price") }}</label>
            <input v-model.number="levelForm.price" class="input" type="number" min="0.01" step="0.01" required />
          </div>
          <div>
            <label class="input-label">{{ t("payment.admin.originalPrice") }}</label>
            <input v-model.number="levelForm.original_price" class="input" type="number" min="0" step="0.01" />
          </div>
          <div>
            <label class="input-label">{{ t("payment.admin.validityDays") }}</label>
            <input v-model.number="levelForm.validity_days" class="input" type="number" min="1" required />
          </div>
          <div>
            <label class="input-label">{{ t("payment.admin.vipDiscountMultiplier") }}</label>
            <input v-model.number="levelForm.discount_multiplier" class="input" type="number" min="0.01" max="1" step="0.01" required />
          </div>
          <div>
            <label class="input-label">{{ t("payment.admin.sortOrder") }}</label>
            <input v-model.number="levelForm.sort_order" class="input" type="number" min="0" />
          </div>
        </div>
        <div>
          <label class="input-label">{{ t("payment.admin.planDescription") }}</label>
          <textarea v-model="levelForm.description" class="input" rows="2" />
        </div>
        <div>
          <label class="input-label">{{ t("payment.admin.features") }}</label>
          <textarea v-model="levelForm.features" class="input" rows="4" :placeholder="t('payment.admin.featuresPlaceholder')" />
          <p class="mt-1 text-xs text-muted">{{ t("payment.admin.featuresHint") }}</p>
        </div>
        <label class="flex items-center gap-3 text-sm text-body">
          <input v-model="levelForm.for_sale" type="checkbox" class="h-4 w-4" />
          {{ t("payment.admin.forSale") }}
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="showLevelDialog = false">{{ t("common.cancel") }}</button>
              <button type="submit" form="vip-level-form" class="btn btn-primary" :disabled="saving">{{ saving ? t("common.saving") : t("common.save") }}</button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="showAssignDialog" :title="t('payment.admin.assignVIPTitle')" width="normal" @close="closeAssignDialog">
      <form id="assign-vip-form" class="space-y-5" @submit.prevent="assignVIP">
        <div>
          <label class="input-label">{{ t("payment.admin.selectUser") }}</label>
          <div class="relative" data-assign-vip-user-search>
            <input
              v-model="userSearchKeyword"
              type="text"
              class="input pr-8"
              :placeholder="t('admin.usage.searchUserPlaceholder')"
              @input="debounceSearchUsers"
              @focus="showUserDropdown = true"
            />
            <button
              v-if="selectedUser"
              type="button"
              class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-soft hover:text-body"
              @click="clearUserSelection"
            >
              <Icon name="x" size="sm" />
            </button>
            <div
              v-if="showUserDropdown && (userSearchResults.length > 0 || userSearchKeyword)"
              class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-hairline bg-canvas shadow-lg"
            >
              <div v-if="userSearchLoading" class="px-4 py-3 text-sm text-muted">
                {{ t("common.loading") }}
              </div>
              <div v-else-if="userSearchResults.length === 0 && userSearchKeyword" class="px-4 py-3 text-sm text-muted">
                {{ t("common.noOptionsFound") }}
              </div>
              <button
                v-for="user in userSearchResults"
                :key="user.id"
                type="button"
                class="w-full px-4 py-2 text-left text-sm hover:bg-surface-card"
                @click="selectUser(user)"
              >
                <span class="font-medium text-ink">{{ user.email }}</span>
                <span class="ml-2 text-muted">#{{ user.id }}</span>
              </button>
            </div>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t("payment.admin.selectVIPLevel") }}</label>
          <Select
            v-model="assignForm.vip_level_id"
            :options="vipLevelOptions"
            :placeholder="t('payment.admin.selectVIPLevel')"
          />
        </div>
        <div>
          <label class="input-label">{{ t("payment.admin.vipDays") }}</label>
          <input v-model.number="assignForm.days" type="number" min="1" max="36500" required class="input" />
        </div>
        <div>
          <label class="input-label">{{ t("payment.admin.assignNotes") }}</label>
          <textarea v-model="assignForm.notes" class="input" rows="3" />
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeAssignDialog">{{ t("common.cancel") }}</button>
          <button type="submit" form="assign-vip-form" class="btn btn-primary" :disabled="assigning">
            {{ assigning ? t("payment.admin.assigningVIP") : t("payment.admin.assignVIP") }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('payment.admin.deleteVIPLevel')"
      :message="t('payment.admin.deleteVIPLevelConfirm')"
      :confirm-text="t('common.delete')"
      danger
      @confirm="deleteLevel"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { adminAPI } from "@/api/admin";
import { adminPaymentAPI } from "@/api/admin/payment";
import { useAppStore } from "@/stores/app";
import { extractI18nErrorMessage } from "@/utils/apiError";
import type { VIPLevel, VIPUserSummary } from "@/types/payment";
import type { SimpleUser } from "@/api/admin/usage";
import type { Column } from "@/components/common/types";
import AppLayout from "@/components/layout/AppLayout.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import ConfirmDialog from "@/components/common/ConfirmDialog.vue";
import DataTable from "@/components/common/DataTable.vue";
import Select from "@/components/common/Select.vue";
import Icon from "@/components/icons/Icon.vue";

const { t } = useI18n();
const appStore = useAppStore();

const levels = ref<VIPLevel[]>([]);
const users = ref<VIPUserSummary[]>([]);
const levelsLoading = ref(false);
const usersLoading = ref(false);
const showLevelDialog = ref(false);
const showAssignDialog = ref(false);
const showDeleteDialog = ref(false);
const saving = ref(false);
const assigning = ref(false);
const editingLevel = ref<VIPLevel | null>(null);
const deletingLevel = ref<VIPLevel | null>(null);
const selectedUser = ref<SimpleUser | null>(null);
const userSearchKeyword = ref("");
const userSearchResults = ref<SimpleUser[]>([]);
const userSearchLoading = ref(false);
const showUserDropdown = ref(false);
let userSearchTimeout: ReturnType<typeof setTimeout> | null = null;

const levelForm = reactive({
  name: "",
  description: "",
  price: 0,
  original_price: null as number | null,
  validity_days: 30,
  discount_multiplier: 1,
  features: "",
  for_sale: true,
  sort_order: 0,
});

const assignForm = reactive({
  user_id: null as number | null,
  vip_level_id: null as number | null,
  days: 30,
  notes: "",
});

const levelColumns = computed((): Column[] => [
  { key: "id", label: "ID" },
  { key: "name", label: t("payment.admin.vipLevelName") },
  { key: "price", label: t("payment.admin.price") },
  { key: "validity_days", label: t("payment.admin.validityDays") },
  { key: "discount_multiplier", label: t("payment.admin.vipDiscountMultiplier") },
  { key: "for_sale", label: t("payment.admin.forSale") },
  { key: "sort_order", label: t("payment.admin.sortOrder") },
  { key: "actions", label: t("common.actions") },
]);

const userColumns = computed((): Column[] => [
  { key: "user", label: t("payment.admin.colUser") },
  { key: "current", label: t("payment.vip.current") },
  { key: "total_savings_usd", label: t("payment.vip.totalSavings") },
]);

const vipLevelOptions = computed(() =>
  levels.value.map((level) => ({
    value: level.id,
    label: `${vipLevelDisplayName(level)} · ${level.validity_days}${t("payment.admin.days")}`,
  })),
);

function vipLevelDisplayName(level: VIPLevel): string {
  const name = String(level.name || "").trim();
  if (!name) return `VIP #${level.id}`;
  if (/^\d+$/.test(name)) return `VIP ${name}`;
  return name;
}

function resetForm(level: VIPLevel | null) {
  editingLevel.value = level;
  levelForm.name = level?.name || "";
  levelForm.description = level?.description || "";
  levelForm.price = level?.price || 0;
  levelForm.original_price = level?.original_price ?? null;
  levelForm.validity_days = level?.validity_days || 30;
  levelForm.discount_multiplier = level?.discount_multiplier || 1;
  levelForm.features = level?.features || "";
  levelForm.for_sale = level?.for_sale ?? true;
  levelForm.sort_order = level?.sort_order || 0;
}

function openLevelDialog(level: VIPLevel | null) {
  resetForm(level);
  showLevelDialog.value = true;
}

function openAssignDialog() {
  assignForm.days = 30;
  assignForm.vip_level_id = levels.value[0]?.id ?? null;
  showAssignDialog.value = true;
}

function closeAssignDialog() {
  showAssignDialog.value = false;
  assignForm.user_id = null;
  assignForm.vip_level_id = null;
  assignForm.days = 30;
  assignForm.notes = "";
  clearUserSelection();
}

function buildPayload(): Record<string, unknown> {
  const originalPrice = Number(levelForm.original_price || 0);
  return {
    name: levelForm.name,
    description: levelForm.description,
    price: levelForm.price,
    original_price: originalPrice > 0 ? originalPrice : undefined,
    clear_original_price: originalPrice <= 0,
    validity_days: levelForm.validity_days,
    discount_multiplier: levelForm.discount_multiplier,
    features: levelForm.features,
    benefits: {},
    for_sale: levelForm.for_sale,
    sort_order: levelForm.sort_order,
  };
}

async function loadLevels() {
  levelsLoading.value = true;
  try {
    const res = await adminPaymentAPI.getVIPLevels();
    levels.value = res.data || [];
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("common.error")));
  } finally {
    levelsLoading.value = false;
  }
}

async function loadUsers() {
  usersLoading.value = true;
  try {
    const res = await adminPaymentAPI.getVIPUsers({ page: 1, page_size: 100 });
    users.value = (res.data.items || []).filter((item) => !!item.current);
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("common.error")));
  } finally {
    usersLoading.value = false;
  }
}

function debounceSearchUsers() {
  if (userSearchTimeout) {
    clearTimeout(userSearchTimeout);
  }
  userSearchTimeout = setTimeout(searchUsers, 300);
}

async function searchUsers() {
  const keyword = userSearchKeyword.value.trim();
  if (selectedUser.value && keyword !== selectedUser.value.email) {
    selectedUser.value = null;
    assignForm.user_id = null;
  }
  if (!keyword) {
    userSearchResults.value = [];
    return;
  }
  userSearchLoading.value = true;
  try {
    userSearchResults.value = await adminAPI.usage.searchUsers(keyword);
  } catch (err: unknown) {
    console.error("Failed to search users:", err);
    userSearchResults.value = [];
  } finally {
    userSearchLoading.value = false;
  }
}

function selectUser(user: SimpleUser) {
  selectedUser.value = user;
  userSearchKeyword.value = user.email;
  userSearchResults.value = [];
  showUserDropdown.value = false;
  assignForm.user_id = user.id;
}

function clearUserSelection() {
  selectedUser.value = null;
  userSearchKeyword.value = "";
  userSearchResults.value = [];
  showUserDropdown.value = false;
  assignForm.user_id = null;
}

async function assignVIP() {
  if (!assignForm.user_id) {
    appStore.showError(t("payment.admin.pleaseSelectUser"));
    return;
  }
  if (!assignForm.vip_level_id) {
    appStore.showError(t("payment.admin.pleaseSelectVIPLevel"));
    return;
  }
  if (!assignForm.days || assignForm.days < 1) {
    appStore.showError(t("payment.admin.vipDaysRequired"));
    return;
  }
  assigning.value = true;
  try {
    await adminPaymentAPI.assignVIP({
      user_id: assignForm.user_id,
      vip_level_id: assignForm.vip_level_id,
      days: assignForm.days,
      notes: assignForm.notes || undefined,
    });
    appStore.showSuccess(t("payment.admin.vipAssigned"));
    closeAssignDialog();
    await loadUsers();
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("payment.admin.failedToAssignVIP")));
  } finally {
    assigning.value = false;
  }
}

async function saveLevel() {
  saving.value = true;
  try {
    if (editingLevel.value) {
      await adminPaymentAPI.updateVIPLevel(editingLevel.value.id, buildPayload());
    } else {
      await adminPaymentAPI.createVIPLevel(buildPayload());
    }
    showLevelDialog.value = false;
    await loadLevels();
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("common.error")));
  } finally {
    saving.value = false;
  }
}

async function toggleForSale(level: VIPLevel) {
  try {
    await adminPaymentAPI.updateVIPLevel(level.id, { for_sale: !level.for_sale });
    level.for_sale = !level.for_sale;
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("common.error")));
  }
}

function confirmDelete(level: VIPLevel) {
  deletingLevel.value = level;
  showDeleteDialog.value = true;
}

async function deleteLevel() {
  if (!deletingLevel.value) return;
  try {
    await adminPaymentAPI.deleteVIPLevel(deletingLevel.value.id);
    showDeleteDialog.value = false;
    deletingLevel.value = null;
    await loadLevels();
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("common.error")));
  }
}

function formatDate(value: string): string {
  return value ? new Date(value).toLocaleString() : "-";
}

onMounted(() => {
  loadLevels();
  loadUsers();
});
</script>
