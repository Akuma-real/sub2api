<template>
  <AppLayout>
    <div class="model-marketplace">
      <section class="marketplace-toolbar">
        <div class="search-box">
          <Icon name="search" size="md" class="text-muted-soft" />
          <input
            v-model="filters.search"
            type="search"
            :placeholder="t('modelMarketplace.searchPlaceholder')"
            class="search-input"
          />
        </div>

        <div v-if="visibleFilterCount > 0" class="filter-grid">
          <label v-if="showPlatformFilter" class="filter-field">
            <span>{{ t("modelMarketplace.filters.platform") }}</span>
            <select v-model="filters.platform">
              <option value="all">{{ t("modelMarketplace.filters.all") }}</option>
              <option v-for="platform in platformOptions" :key="platform">
                {{ platform }}
              </option>
            </select>
          </label>

          <label v-if="showGroupFilter" class="filter-field">
            <span>{{ t("modelMarketplace.filters.group") }}</span>
            <select v-model="filters.group">
              <option value="all">{{ t("modelMarketplace.filters.all") }}</option>
              <option v-for="group in groupOptions" :key="group">
                {{ group }}
              </option>
            </select>
          </label>

          <label v-if="showBillingFilter" class="filter-field">
            <span>{{ t("modelMarketplace.filters.billing") }}</span>
            <select v-model="filters.billing">
              <option value="all">{{ t("modelMarketplace.filters.all") }}</option>
              <option v-for="mode in billingModeOptions" :key="mode" :value="mode">
                {{ billingModeLabel(mode) }}
              </option>
            </select>
          </label>

          <label v-if="showPricingFilter" class="filter-field">
            <span>{{ t("modelMarketplace.filters.pricing") }}</span>
            <select v-model="filters.pricing">
              <option value="all">{{ t("modelMarketplace.filters.all") }}</option>
              <option value="priced">
                {{ t("modelMarketplace.filters.priced") }}
              </option>
              <option value="unpriced">
                {{ t("modelMarketplace.filters.unpriced") }}
              </option>
            </select>
          </label>

        </div>

        <div class="toolbar-actions">
          <label class="sort-field">
            <span>{{ t("modelMarketplace.sort.label") }}</span>
            <select v-model="sortKey">
              <option value="name">{{ t("modelMarketplace.sort.name") }}</option>
              <option value="platform">
                {{ t("modelMarketplace.sort.platform") }}
              </option>
              <option value="input">
                {{ t("modelMarketplace.sort.input") }}
              </option>
              <option value="output">
                {{ t("modelMarketplace.sort.output") }}
              </option>
              <option value="rate">
                {{ t("modelMarketplace.sort.rate") }}
              </option>
            </select>
          </label>
          <div class="view-toggle">
            <button
              type="button"
              :class="{ active: viewMode === 'table' }"
              @click="viewMode = 'table'"
            >
              <Icon name="sort" size="sm" />
              <span>{{ t("modelMarketplace.view.table") }}</span>
            </button>
            <button
              type="button"
              :class="{ active: viewMode === 'cards' }"
              @click="viewMode = 'cards'"
            >
              <Icon name="grid" size="sm" />
              <span>{{ t("modelMarketplace.view.cards") }}</span>
            </button>
          </div>
        </div>
      </section>

      <section class="result-meta">
        <p>
          {{
            t("modelMarketplace.resultCount", {
              count: filteredModels.length,
              total: models.length,
            })
          }}
        </p>
        <button type="button" class="text-link" @click="resetFilters">
          {{ t("modelMarketplace.resetFilters") }}
        </button>
      </section>

      <section v-if="loading" class="loading-list">
        <div v-for="i in 6" :key="i" class="loading-row"></div>
      </section>

      <section v-else-if="filteredModels.length === 0" class="empty-state">
        <Icon name="inbox" size="xl" />
        <p>{{ t("modelMarketplace.empty") }}</p>
      </section>

      <section
        v-else-if="viewMode === 'cards'"
        class="model-card-grid"
        aria-label="Model marketplace cards"
      >
        <article
          v-for="model in filteredModels"
          :key="`${model.platform}:${model.id}`"
          class="model-card"
        >
          <div class="model-card-header">
            <div>
              <p class="model-platform">
                <PlatformIcon :platform="model.platform as GroupPlatform" size="xs" />
                {{ model.platform }}
              </p>
              <h2>{{ model.display_name }}</h2>
            </div>
            <button
              type="button"
              class="icon-button"
              :title="t('modelMarketplace.copyModelId')"
              @click="copyModelId(model.id)"
            >
              <Icon name="copy" size="sm" />
            </button>
          </div>
          <div class="model-card-body">
            <div class="price-line">
              <span>{{ t("modelMarketplace.columns.price") }}</span>
              <strong>{{ priceSummary(model) }}</strong>
            </div>
            <div class="price-line">
              <span>{{ t("modelMarketplace.columns.rate") }}</span>
              <strong>{{ modelRateSummary(model) }}</strong>
            </div>
            <div class="model-badges">
              <span>{{ billingModeLabel(model.billing_mode) }}</span>
              <span>{{ pricingSourceLabel(model.pricing_source) }}</span>
            </div>
            <div class="group-chip-list">
              <span
                v-for="group in modelGroupNames(model)"
                :key="group"
                class="group-chip"
              >
                {{ group }}
              </span>
            </div>
          </div>
          <div class="model-card-actions">
            <button type="button" class="btn btn-secondary" @click="openDetail(model)">
              <Icon name="eye" size="sm" />
              <span>{{ t("modelMarketplace.viewDetails") }}</span>
            </button>
          </div>
        </article>
      </section>

      <section v-else class="model-table-wrap">
        <table class="model-table">
          <thead>
            <tr>
              <th>{{ t("modelMarketplace.columns.model") }}</th>
              <th>{{ t("modelMarketplace.columns.platform") }}</th>
              <th>{{ t("modelMarketplace.columns.price") }}</th>
              <th>{{ t("modelMarketplace.columns.rate") }}</th>
              <th>{{ t("modelMarketplace.columns.groups") }}</th>
              <th>{{ t("modelMarketplace.columns.actions") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="model in filteredModels"
              :key="`${model.platform}:${model.id}`"
            >
              <td>
                <div class="model-cell">
                  <button
                    type="button"
                    class="model-id-button"
                    @click="openDetail(model)"
                  >
                    {{ model.display_name }}
                  </button>
                  <span>{{ billingModeLabel(model.billing_mode) }}</span>
                </div>
              </td>
              <td>
                <span class="platform-pill">
                  <PlatformIcon :platform="model.platform as GroupPlatform" size="xs" />
                  {{ model.platform }}
                </span>
              </td>
              <td>
                <div class="price-cell">
                  <strong>{{ priceSummary(model) }}</strong>
                  <span>{{ pricingSourceLabel(model.pricing_source) }}</span>
                </div>
              </td>
              <td>
                <div class="rate-cell">
                  <strong>{{ modelRateSummary(model) }}</strong>
                  <span>{{ t("modelMarketplace.groups.effectiveRateLabel") }}</span>
                </div>
              </td>
              <td>
                <div class="group-chip-list">
                  <span
                    v-for="group in modelGroupNames(model)"
                    :key="group"
                    class="group-chip"
                  >
                    {{ group }}
                  </span>
                </div>
              </td>
              <td>
                <div class="row-actions">
                  <button
                    type="button"
                    class="icon-button"
                    :title="t('modelMarketplace.copyModelId')"
                    @click="copyModelId(model.id)"
                  >
                    <Icon name="copy" size="sm" />
                  </button>
                  <button
                    type="button"
                    class="btn btn-secondary"
                    @click="openDetail(model)"
                  >
                    <Icon name="eye" size="sm" />
                    <span>{{ t("modelMarketplace.details") }}</span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </section>
    </div>

    <BaseDialog
      :show="selectedModel != null"
      :title="selectedModel?.display_name || ''"
      width="extra-wide"
      @close="selectedModel = null"
    >
      <div v-if="selectedModel" class="detail-panel">
        <div class="detail-summary">
          <div>
            <p class="model-platform">
              <PlatformIcon
                :platform="selectedModel.platform as GroupPlatform"
                size="xs"
              />
              {{ selectedModel.platform }}
            </p>
            <p class="detail-model-id">{{ selectedModel.id }}</p>
          </div>
          <button
            type="button"
            class="btn btn-primary"
            @click="copyModelId(selectedModel.id)"
          >
            <Icon name="copy" size="sm" />
            <span>{{ t("modelMarketplace.copyModelId") }}</span>
          </button>
        </div>

        <div class="detail-grid">
          <div class="detail-metric">
            <span>{{ t("modelMarketplace.columns.price") }}</span>
            <strong>{{ priceSummary(selectedModel) }}</strong>
          </div>
          <div class="detail-metric">
            <span>{{ t("modelMarketplace.columns.rate") }}</span>
            <strong>{{ modelRateSummary(selectedModel) }}</strong>
          </div>
          <div class="detail-metric">
            <span>{{ t("modelMarketplace.columns.groups") }}</span>
            <div class="group-chip-list">
              <span
                v-for="group in modelGroupNames(selectedModel)"
                :key="group"
                class="group-chip"
              >
                {{ group }}
              </span>
            </div>
          </div>
        </div>

        <div class="request-sample">
          <div class="request-sample-header">
            <span>{{ t("modelMarketplace.requestExample") }}</span>
            <button
              type="button"
              class="icon-button on-dark"
              :title="t('common.copy', 'Copy')"
              @click="copyRequestExample(selectedModel.id)"
            >
              <Icon name="copy" size="sm" />
            </button>
          </div>
          <pre><code>{{ requestExample(selectedModel.id) }}</code></pre>
        </div>

        <div class="group-detail-list">
          <article
            v-for="detail in modelGroupDetails(selectedModel)"
            :key="detail.group.id"
            class="group-detail"
          >
            <div class="group-detail-header">
              <div>
                <h3>{{ detail.group.name }}</h3>
                <p>
                  {{
                    detail.group.is_exclusive
                      ? t("modelMarketplace.groups.exclusive")
                      : t("modelMarketplace.groups.public")
                  }}
                </p>
              </div>
              <span>{{ pricingSourceLabel(detail.pricingSource) }}</span>
            </div>

            <div class="group-rate-summary">
              <strong>
                {{
                  t("modelMarketplace.groups.effectiveRate", {
                    rate: formatMultiplier(detail.group.effective_rate_multiplier),
                  })
                }}
              </strong>
            </div>

            <div class="pricing-grid">
              <div>
                <span>{{ t("modelMarketplace.pricing.input") }}</span>
                <strong>{{ formatMillion(detail.pricing?.input_price ?? null) }}</strong>
              </div>
              <div>
                <span>{{ t("modelMarketplace.pricing.output") }}</span>
                <strong>{{ formatMillion(detail.pricing?.output_price ?? null) }}</strong>
              </div>
              <div>
                <span>{{ t("modelMarketplace.pricing.cacheWrite") }}</span>
                <strong>{{
                  formatMillion(detail.pricing?.cache_write_price ?? null)
                }}</strong>
              </div>
              <div>
                <span>{{ t("modelMarketplace.pricing.cacheRead") }}</span>
                <strong>{{
                  formatMillion(detail.pricing?.cache_read_price ?? null)
                }}</strong>
              </div>
            </div>
          </article>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import AppLayout from "@/components/layout/AppLayout.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import Icon from "@/components/icons/Icon.vue";
import PlatformIcon from "@/components/common/PlatformIcon.vue";
import type { GroupPlatform } from "@/types";
import modelMarketplaceAPI, {
  type ModelMarketplaceModel,
  type ModelMarketplaceResponse,
} from "@/api/modelMarketplace";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage } from "@/utils/apiError";
import { useClipboard } from "@/composables/useClipboard";
import { formatScaled } from "@/utils/pricing";
import {
  BILLING_MODE_IMAGE,
  BILLING_MODE_PER_REQUEST,
  BILLING_MODE_TOKEN,
  type BillingMode,
} from "@/constants/channel";

type SortKey = "name" | "platform" | "input" | "output" | "rate";
type ViewMode = "table" | "cards";
type ModelMarketplaceChannel = ModelMarketplaceModel["channels"][number];
type ModelMarketplaceGroup = ModelMarketplaceChannel["groups"][number];
type ModelMarketplacePricing = ModelMarketplaceChannel["pricing"];

interface ModelGroupDetail {
  group: ModelMarketplaceGroup;
  pricing: ModelMarketplacePricing;
  pricingSource: ModelMarketplaceChannel["pricing_source"];
}

const { t } = useI18n();
const appStore = useAppStore();
const { copyToClipboard } = useClipboard();

const response = ref<ModelMarketplaceResponse>({
  summary: {
    models: 0,
    platforms: 0,
    channels: 0,
    groups: 0,
    priced_models: 0,
    unpriced_models: 0,
    price_variants: 0,
  },
  models: [],
});
const loading = ref(false);
const selectedModel = ref<ModelMarketplaceModel | null>(null);
const sortKey = ref<SortKey>("name");
const viewMode = ref<ViewMode>("table");
const filters = reactive({
  search: "",
  platform: "all",
  group: "all",
  billing: "all",
  pricing: "all",
});

const models = computed(() => response.value.models);

const platformOptions = computed(() =>
  Array.from(new Set(models.value.map((m) => m.platform))).sort(),
);

const groupOptions = computed(() =>
  Array.from(
    new Set(
      models.value.flatMap((m) =>
        m.channels.flatMap((c) => c.groups.map((g) => g.name)),
      ),
    ),
  ).sort((a, b) => a.localeCompare(b)),
);

const billingModeOptions = computed(() =>
  Array.from(new Set(models.value.map((m) => m.billing_mode))).sort(),
);

const showPlatformFilter = computed(() => platformOptions.value.length > 1);
const showGroupFilter = computed(() => groupOptions.value.length > 1);
const showBillingFilter = computed(() => billingModeOptions.value.length > 1);
const showPricingFilter = computed(() => {
  const hasPriced = models.value.some((model) => Boolean(model.pricing));
  const hasUnpriced = models.value.some((model) => !model.pricing);
  return hasPriced && hasUnpriced;
});
const visibleFilterCount = computed(
  () =>
    Number(showPlatformFilter.value) +
    Number(showGroupFilter.value) +
    Number(showBillingFilter.value) +
    Number(showPricingFilter.value),
);

const filteredModels = computed(() => {
  const q = filters.search.trim().toLowerCase();
  const rows = models.value.filter((model) => {
    if (
      showPlatformFilter.value &&
      filters.platform !== "all" &&
      model.platform !== filters.platform
    ) {
      return false;
    }
    if (
      showBillingFilter.value &&
      filters.billing !== "all" &&
      model.billing_mode !== filters.billing
    ) {
      return false;
    }
    if (showPricingFilter.value) {
      if (filters.pricing === "priced" && !model.pricing) return false;
      if (filters.pricing === "unpriced" && model.pricing) return false;
    }
    if (
      showGroupFilter.value &&
      filters.group !== "all" &&
      !model.channels.some((c) =>
        c.groups.some((g) => g.name === filters.group),
      )
    ) {
      return false;
    }
    if (!q) return true;
    return (
      model.id.toLowerCase().includes(q) ||
      model.platform.toLowerCase().includes(q) ||
      model.channels.some((channel) =>
        channel.groups.some((g) => g.name.toLowerCase().includes(q)),
      )
    );
  });

  return [...rows].sort((a, b) => compareModels(a, b, sortKey.value));
});

function compareModels(
  a: ModelMarketplaceModel,
  b: ModelMarketplaceModel,
  key: SortKey,
) {
  switch (key) {
    case "platform":
      return a.platform.localeCompare(b.platform) || a.id.localeCompare(b.id);
    case "input":
      return priceForSort(a, "input_price") - priceForSort(b, "input_price");
    case "output":
      return priceForSort(a, "output_price") - priceForSort(b, "output_price");
    case "rate":
      return rateForSort(a) - rateForSort(b) || a.id.localeCompare(b.id);
    default:
      return a.id.localeCompare(b.id);
  }
}

function priceForSort(
  model: ModelMarketplaceModel,
  field: "input_price" | "output_price",
) {
  const value = model.pricing?.[field];
  return value == null ? Number.POSITIVE_INFINITY : value;
}

function modelGroupNames(model: ModelMarketplaceModel) {
  return Array.from(
    new Set(
      model.channels.flatMap((channel) =>
        channel.groups.map((group) => group.name),
      ),
    ),
  ).sort((a, b) => a.localeCompare(b));
}

function modelGroupDetails(model: ModelMarketplaceModel): ModelGroupDetail[] {
  const groupMap = new Map<
    number,
    {
      group: ModelMarketplaceGroup;
      pricing: ModelMarketplacePricing;
      pricingSource: ModelMarketplaceChannel["pricing_source"];
    }
  >();

  for (const channel of model.channels) {
    for (const group of channel.groups) {
      const existing = groupMap.get(group.id);
      if (!existing) {
        groupMap.set(group.id, {
          group,
          pricing: channel.pricing,
          pricingSource: channel.pricing_source,
        });
        continue;
      }

      if (!existing.pricing && channel.pricing) {
        existing.pricing = channel.pricing;
        existing.pricingSource = channel.pricing_source;
      }
    }
  }

  return Array.from(groupMap.values())
    .map((detail) => ({
      group: detail.group,
      pricing: detail.pricing,
      pricingSource: detail.pricingSource,
    }))
    .sort((a, b) => a.group.name.localeCompare(b.group.name));
}

function modelRateValues(model: ModelMarketplaceModel) {
  const rates = model.channels.flatMap((channel) =>
    channel.groups
      .map((group) => group.effective_rate_multiplier)
      .filter((rate): rate is number => Number.isFinite(rate)),
  );

  return Array.from(new Set(rates.map((rate) => Number(rate.toFixed(6))))).sort(
    (a, b) => a - b,
  );
}

function modelRateSummary(model: ModelMarketplaceModel) {
  const rates = modelRateValues(model);
  if (rates.length === 0) return "-";
  if (rates.length === 1) return formatMultiplier(rates[0]);
  return `${formatMultiplier(rates[0])} - ${formatMultiplier(
    rates[rates.length - 1],
  )}`;
}

function rateForSort(model: ModelMarketplaceModel) {
  return modelRateValues(model)[0] ?? Number.POSITIVE_INFINITY;
}

function resetFilters() {
  filters.search = "";
  filters.platform = "all";
  filters.group = "all";
  filters.billing = "all";
  filters.pricing = "all";
  sortKey.value = "name";
}

function openDetail(model: ModelMarketplaceModel) {
  selectedModel.value = model;
}

function copyModelId(model: string) {
  copyToClipboard(model, t("modelMarketplace.modelCopied"));
}

function copyRequestExample(model: string) {
  copyToClipboard(requestExample(model), t("modelMarketplace.requestCopied"));
}

function requestExample(model: string) {
  return `curl "$BASE_URL/v1/chat/completions" \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"${model}","messages":[{"role":"user","content":"Hello"}]}'`;
}

function billingModeLabel(mode: BillingMode | string) {
  switch (mode) {
    case BILLING_MODE_PER_REQUEST:
      return t("modelMarketplace.billing.perRequest");
    case BILLING_MODE_IMAGE:
      return t("modelMarketplace.billing.image");
    case BILLING_MODE_TOKEN:
    default:
      return t("modelMarketplace.billing.token");
  }
}

function pricingSourceLabel(source: string) {
  switch (source) {
    case "channel":
      return t("modelMarketplace.pricingSource.channel");
    case "global":
      return t("modelMarketplace.pricingSource.global");
    default:
      return t("modelMarketplace.pricingSource.none");
  }
}

function priceSummary(model: ModelMarketplaceModel) {
  const pricing = model.pricing;
  if (!pricing) return t("modelMarketplace.noPricing");
  if (pricing.billing_mode === BILLING_MODE_PER_REQUEST) {
    return `${formatScaled(pricing.per_request_price, 1)} ${t(
      "modelMarketplace.pricing.perRequestUnit",
    )}`;
  }
  if (pricing.billing_mode === BILLING_MODE_IMAGE) {
    return `${formatScaled(pricing.image_output_price, 1)} ${t(
      "modelMarketplace.pricing.perImageUnit",
    )}`;
  }
  return `${formatMillion(pricing.input_price)} / ${formatMillion(
    pricing.output_price,
  )}`;
}

function formatMillion(value: number | null) {
  return `${formatScaled(value, 1_000_000)} ${t(
    "modelMarketplace.pricing.perMillion",
  )}`;
}

function formatMultiplier(value: number | null) {
  if (value == null) return "-";
  return `${value.toFixed(3).replace(/\.?0+$/, "")}x`;
}

async function loadMarketplace() {
  loading.value = true;
  try {
    response.value = await modelMarketplaceAPI.getMarketplace();
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t("common.error")));
  } finally {
    loading.value = false;
  }
}

onMounted(loadMarketplace);
</script>

<style scoped>
.model-marketplace {
  @apply mx-auto flex w-full max-w-7xl flex-col gap-6;
}

.marketplace-toolbar {
  @apply flex flex-col gap-4 rounded-lg border border-hairline bg-surface-soft p-4;
}

.search-box {
  @apply flex h-11 items-center gap-2 rounded-md border border-hairline bg-canvas px-3;
}

.search-input {
  @apply h-full flex-1 bg-transparent text-sm text-ink outline-none placeholder:text-muted-soft;
}

.filter-grid {
  @apply grid gap-3 md:grid-cols-2 xl:grid-cols-3;
}

.filter-field,
.sort-field {
  @apply flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-muted;
}

.filter-field select,
.sort-field select {
  @apply h-10 rounded-md border border-hairline bg-canvas py-0 pl-3 pr-10 text-sm normal-case tracking-normal text-ink outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/15;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%23756f68' stroke-width='2' stroke-linecap='round' stroke-linejoin='round' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
  background-position: right 1rem center;
  background-repeat: no-repeat;
  background-size: 0.85rem;
}

.toolbar-actions {
  @apply flex flex-col justify-between gap-3 border-t border-hairline pt-4 md:flex-row md:items-end;
}

.view-toggle {
  @apply inline-flex rounded-md border border-hairline bg-canvas p-1;
}

.view-toggle button {
  @apply inline-flex h-9 items-center gap-2 rounded-md px-3 text-sm font-medium text-muted;
}

.view-toggle button.active {
  @apply bg-surface-card text-ink;
}

.result-meta {
  @apply flex items-center justify-between text-sm text-muted;
}

.text-link {
  @apply text-sm font-medium text-primary-700 hover:text-primary-800;
}

.loading-list {
  @apply space-y-2;
}

.loading-row {
  @apply h-16 animate-pulse rounded-lg border border-hairline bg-canvas;
}

.empty-state {
  @apply flex flex-col items-center justify-center gap-3 rounded-lg border border-hairline bg-canvas py-16 text-muted;
}

.model-card-grid {
  @apply grid gap-4 md:grid-cols-2 xl:grid-cols-3;
}

.model-card {
  @apply flex flex-col gap-4 rounded-lg border border-hairline bg-canvas p-5;
}

.model-card-header {
  @apply flex items-start justify-between gap-3;
}

.model-card h2 {
  @apply mt-2 break-all font-serif text-2xl font-normal leading-tight tracking-[-0.01em] text-ink;
}

.model-platform {
  @apply inline-flex items-center gap-1 text-xs font-medium uppercase tracking-wide text-muted;
}

.model-card-body {
  @apply space-y-3 border-t border-hairline pt-4;
}

.price-line {
  @apply flex items-center justify-between gap-4 text-sm text-muted;
}

.price-line strong {
  @apply text-right font-medium text-ink;
}

.model-badges {
  @apply flex flex-wrap gap-1.5;
}

.model-badges span,
.platform-pill {
  @apply inline-flex items-center gap-1 rounded-full bg-surface-card px-2.5 py-1 text-xs font-medium text-body;
}

.group-chip-list {
  @apply flex flex-wrap gap-1.5;
}

.group-chip {
  @apply inline-flex max-w-[12rem] items-center truncate rounded-full bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700;
}

.model-card-actions {
  @apply mt-auto flex justify-end;
}

.model-table-wrap {
  @apply overflow-x-auto rounded-lg border border-hairline bg-canvas;
}

.model-table {
  @apply w-full min-w-[960px] border-collapse text-sm;
}

.model-table th {
  @apply border-b border-hairline bg-surface-soft px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-muted;
}

.model-table td {
  @apply border-b border-hairline-soft px-4 py-4 align-top text-body;
}

.model-table tbody tr:last-child td {
  @apply border-b-0;
}

.model-cell,
.price-cell,
.rate-cell {
  @apply flex flex-col gap-1;
}

.model-id-button {
  @apply break-all text-left font-medium text-ink hover:text-primary-700;
}

.model-cell span,
.price-cell span,
.rate-cell span {
  @apply text-xs text-muted;
}

.rate-cell strong {
  @apply font-medium text-ink;
}

.row-actions {
  @apply flex items-center gap-2;
}

.icon-button {
  @apply inline-flex h-9 w-9 items-center justify-center rounded-md border border-hairline bg-canvas text-muted transition-colors hover:bg-surface-card hover:text-ink;
}

.icon-button.on-dark {
  @apply border-surface-dark-elevated bg-surface-dark-elevated text-on-dark hover:bg-surface-dark-soft;
}

.detail-panel {
  @apply space-y-6;
}

.detail-summary {
  @apply flex flex-col justify-between gap-4 border-b border-hairline pb-5 md:flex-row md:items-center;
}

.detail-model-id {
  @apply mt-2 break-all font-serif text-3xl font-normal leading-tight text-ink;
}

.detail-grid {
  @apply grid gap-3 md:grid-cols-2;
}

.detail-metric {
  @apply rounded-lg border border-hairline bg-canvas p-4;
}

.detail-metric > span {
  @apply text-xs font-medium uppercase tracking-wide text-muted;
}

.detail-metric strong {
  @apply mt-2 block text-lg font-medium text-ink;
}

.request-sample {
  @apply overflow-hidden rounded-lg bg-surface-dark text-on-dark;
}

.request-sample-header {
  @apply flex items-center justify-between border-b border-surface-dark-elevated px-4 py-3 text-sm font-medium;
}

.request-sample pre {
  @apply overflow-x-auto p-4 font-mono text-sm leading-6 text-on-dark-soft;
}

.group-detail-list {
  @apply space-y-4;
}

.group-detail {
  @apply rounded-lg border border-hairline bg-canvas p-5;
}

.group-detail-header {
  @apply flex flex-col justify-between gap-3 md:flex-row md:items-start;
}

.group-detail-header h3 {
  @apply text-lg font-medium text-ink;
}

.group-detail-header p {
  @apply mt-1 text-sm text-muted;
}

.group-detail-header > span {
  @apply inline-flex w-fit rounded-full bg-surface-card px-3 py-1 text-xs font-medium text-body;
}

.group-rate-summary {
  @apply mt-4 flex flex-wrap items-center gap-2 text-sm;
}

.group-rate-summary span {
  @apply inline-flex rounded-full bg-surface-card px-3 py-1 text-xs font-medium text-body;
}

.group-rate-summary strong {
  @apply inline-flex rounded-full bg-primary-50 px-3 py-1 text-xs font-semibold text-primary-700;
}

.pricing-grid {
  @apply mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4;
}

.pricing-grid div {
  @apply rounded-md border border-hairline bg-surface-soft p-3;
}

.pricing-grid span {
  @apply block text-xs text-muted;
}

.pricing-grid strong {
  @apply mt-1 block text-sm font-medium text-ink;
}

</style>
