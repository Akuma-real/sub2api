<template>
  <BaseDialog
    :show="show"
    :title="group ? t('admin.groups.accounts.manageTitle', { name: group.name }) : ''"
    width="wide"
    @close="$emit('close')"
  >
    <div class="space-y-4">
      <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_12rem_12rem]">
        <label class="block">
          <span class="input-label">{{ t("common.search") }}</span>
          <input
            v-model="search"
            class="input"
            :placeholder="t('admin.groups.accounts.searchPlaceholder')"
            @keyup.enter="loadAccounts"
          />
        </label>
        <label class="block">
          <span class="input-label">{{ t("admin.groups.form.platform") }}</span>
          <Select
            v-model="platform"
            class="w-full"
            :options="platformOptions"
            @change="loadAccounts"
          />
        </label>
        <label class="block">
          <span class="input-label">{{ t("admin.groups.allStatus") }}</span>
          <Select
            v-model="status"
            class="w-full"
            :options="statusOptions"
            @change="loadAccounts"
          />
        </label>
      </div>

      <div class="overflow-hidden rounded-lg border border-hairline">
        <div
          class="grid grid-cols-[3rem_minmax(0,1fr)_8rem_7rem] bg-surface-soft px-3 py-2 text-xs font-medium uppercase text-muted"
        >
          <span></span>
          <span>{{ t("admin.accounts.columns.name") }}</span>
          <span>{{ t("admin.accounts.columns.platform") }}</span>
          <span>{{ t("admin.accounts.columns.status") }}</span>
        </div>
        <div v-if="loading" class="p-6 text-center text-sm text-muted">
          {{ t("common.loading") }}
        </div>
        <template v-else>
          <label
            v-for="account in accounts"
            :key="account.id"
            class="grid cursor-pointer grid-cols-[3rem_minmax(0,1fr)_8rem_7rem] items-center border-t border-hairline px-3 py-2 text-sm hover:bg-surface-card"
          >
            <input
              type="checkbox"
              class="h-4 w-4"
              :checked="selectedIds.has(account.id)"
              @change="toggleAccount(account.id)"
            />
            <span class="min-w-0 truncate text-body">{{ account.name }}</span>
            <span class="text-muted">{{ account.platform }}</span>
            <span class="text-muted">{{
              t(`admin.accounts.status.${account.status}`)
            }}</span>
          </label>
        </template>
        <div
          v-if="!loading && accounts.length === 0"
          class="p-6 text-center text-sm text-muted"
        >
          {{ t("common.noData") }}
        </div>
      </div>

      <div class="flex items-center justify-between text-sm text-muted">
        <span>
          {{
            t("admin.groups.accounts.selectedCount", {
              count: selectedIds.size,
            })
          }}
        </span>
        <button class="btn btn-secondary btn-sm" @click="loadAccounts">
          <Icon name="refresh" size="sm" />
          {{ t("common.refresh") }}
        </button>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3 pt-4">
        <button class="btn btn-secondary" @click="$emit('close')">
          {{ t("common.cancel") }}
        </button>
        <button class="btn btn-primary" :disabled="saving" @click="save">
          {{ saving ? t("common.saving") : t("common.save") }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { adminAPI } from "@/api/admin";
import BaseDialog from "@/components/common/BaseDialog.vue";
import Select from "@/components/common/Select.vue";
import Icon from "@/components/icons/Icon.vue";
import { useAppStore } from "@/stores/app";
import type { Account, AdminGroup } from "@/types";

const props = defineProps<{
  show: boolean;
  group: AdminGroup | null;
}>();

const emit = defineEmits<{
  close: [];
  saved: [];
}>();

const { t } = useI18n();
const appStore = useAppStore();
const accounts = ref<Account[]>([]);
const loadedAccountsById = ref<Map<number, Account>>(new Map());
const knownAccountIds = ref<Set<number>>(new Set());
const originalIds = ref<Set<number>>(new Set());
const selectedIds = ref<Set<number>>(new Set());
const search = ref("");
const platform = ref("");
const status = ref("");
const loading = ref(false);
const saving = ref(false);
const platformOptions = computed(() => [
  { value: "", label: t("admin.groups.allPlatforms") },
  { value: "anthropic", label: "Anthropic" },
  { value: "openai", label: "OpenAI" },
  { value: "gemini", label: "Gemini" },
  { value: "antigravity", label: "Antigravity" },
]);
const statusOptions = computed(() => [
  { value: "", label: t("admin.groups.allStatus") },
  { value: "active", label: t("admin.accounts.status.active") },
  { value: "inactive", label: t("admin.accounts.status.inactive") },
  { value: "error", label: t("admin.accounts.status.error") },
]);

function accountGroupIds(account: Account): number[] {
  return account.group_ids ?? account.groups?.map((group) => group.id) ?? [];
}

async function loadAccounts() {
  if (!props.group) return;
  const groupId = props.group.id;
  loading.value = true;
  try {
    const result = await adminAPI.accounts.list(1, 500, {
      search: search.value,
      platform: platform.value,
      status: status.value,
    });
    accounts.value = result.items;
    const nextKnown = new Set(knownAccountIds.value);
    const nextOriginal = new Set(originalIds.value);
    const nextSelected = new Set(selectedIds.value);
    const nextLoaded = new Map(loadedAccountsById.value);
    for (const account of result.items) {
      nextLoaded.set(account.id, account);
      if (nextKnown.has(account.id)) continue;
      nextKnown.add(account.id);
      if (accountGroupIds(account).includes(groupId)) {
        nextOriginal.add(account.id);
        nextSelected.add(account.id);
      }
    }
    knownAccountIds.value = nextKnown;
    originalIds.value = nextOriginal;
    selectedIds.value = nextSelected;
    loadedAccountsById.value = nextLoaded;
  } catch (error) {
    console.error("Failed to load group accounts: ", error);
    appStore.showError(t("admin.groups.accounts.failedToLoad"));
  } finally {
    loading.value = false;
  }
}

function toggleAccount(id: number) {
  const next = new Set(selectedIds.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  selectedIds.value = next;
}

async function save() {
  if (!props.group || saving.value) return;
  saving.value = true;
  try {
    const groupId = props.group.id;
    for (const account of loadedAccountsById.value.values()) {
      const wasSelected = originalIds.value.has(account.id);
      const isSelected = selectedIds.value.has(account.id);
      if (wasSelected === isSelected) continue;
      const current = accountGroupIds(account);
      const group_ids = isSelected
        ? [...new Set([...current, groupId])]
        : current.filter((id) => id !== groupId);
      await adminAPI.accounts.update(account.id, { group_ids });
    }
    appStore.showSuccess(t("common.success"));
    emit("saved");
  } catch (error: any) {
    console.error("Failed to save group accounts: ", error);
    appStore.showError(error?.response?.data?.message || t("common.error"));
  } finally {
    saving.value = false;
  }
}

watch(
  () => props.show,
	  (show) => {
	    if (!show) return;
	    search.value = "";
	    platform.value = props.group?.platform || "";
	    status.value = "";
	    loadedAccountsById.value = new Map();
	    knownAccountIds.value = new Set();
	    originalIds.value = new Set();
	    selectedIds.value = new Set();
	    loadAccounts();
	  },
	);
</script>
