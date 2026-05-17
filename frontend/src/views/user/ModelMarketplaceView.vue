<template>
  <AppLayout>
    <div class="model-marketplace">
      <section class="summary-grid" aria-label="Model marketplace summary">
        <div
          v-for="card in summaryCards"
          :key="card.key"
          class="summary-card"
        >
          <div class="summary-icon">
            <Icon :name="card.icon" size="md" />
          </div>
          <div>
            <p class="summary-label">{{ card.label }}</p>
            <p class="summary-value">{{ card.value }}</p>
          </div>
        </div>
      </section>

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

        <div class="filter-grid">
          <label class="filter-field">
            <span>{{ t("modelMarketplace.filters.platform") }}</span>
            <select v-model="filters.platform">
              <option value="all">{{ t("modelMarketplace.filters.all") }}</option>
              <option v-for="platform in platformOptions" :key="platform">
                {{ platform }}
              </option>
            </select>
          </label>

          <label class="filter-field">
            <span>{{ t("modelMarketplace.filters.channel") }}</span>
            <select v-model="filters.channel">
              <option value="all">{{ t("modelMarketplace.filters.all") }}</option>
              <option v-for="channel in channelOptions" :key="channel">
                {{ channel }}
              </option>
            </select>
          </label>

          <label class="filter-field">
            <span>{{ t("modelMarketplace.filters.group") }}</span>
            <select v-model="filters.group">
              <option value="all">{{ t("modelMarketplace.filters.all") }}</option>
              <option v-for="group in groupOptions" :key="group">
                {{ group }}
              </option>
            </select>
          </label>

          <label class="filter-field">
            <span>{{ t("modelMarketplace.filters.billing") }}</span>
            <select v-model="filters.billing">
              <option value="all">{{ t("modelMarketplace.filters.all") }}</option>
              <option value="token">
                {{ t("modelMarketplace.billing.token") }}
              </option>
              <option value="per_request">
                {{ t("modelMarketplace.billing.perRequest") }}
              </option>
              <option value="image">
                {{ t("modelMarketplace.billing.image") }}
              </option>
            </select>
          </label>

          <label class="filter-field">
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

          <label class="filter-field">
            <span>{{ t("modelMarketplace.filters.capability") }}</span>
            <select v-model="filters.capability">
              <option value="all">{{ t("modelMarketplace.filters.all") }}</option>
              <option value="image">
                {{ t("modelMarketplace.capabilities.image") }}
              </option>
              <option value="cache">
                {{ t("modelMarketplace.capabilities.cache") }}
              </option>
              <option value="tiered">
                {{ t("modelMarketplace.capabilities.tiered") }}
              </option>
              <option value="per_request">
                {{ t("modelMarketplace.capabilities.perRequest") }}
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
              <option value="channels">
                {{ t("modelMarketplace.sort.channels") }}
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
            <div class="model-badges">
              <span>{{ billingModeLabel(model.billing_mode) }}</span>
              <span>{{ pricingSourceLabel(model.pricing_source) }}</span>
              <span>{{
                t("modelMarketplace.channelCount", {
                  count: model.channel_count,
                })
              }}</span>
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
              <th>{{ t("modelMarketplace.columns.channels") }}</th>
              <th>{{ t("modelMarketplace.columns.groups") }}</th>
              <th>{{ t("modelMarketplace.columns.capabilities") }}</th>
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
              <td>{{ model.channel_count }}</td>
              <td>{{ model.group_count }}</td>
              <td>
                <div class="capability-list">
                  <span v-if="model.capabilities.supports_image">
                    {{ t("modelMarketplace.capabilities.image") }}
                  </span>
                  <span v-if="model.capabilities.supports_cache_pricing">
                    {{ t("modelMarketplace.capabilities.cache") }}
                  </span>
                  <span v-if="model.capabilities.has_tiered_pricing">
                    {{ t("modelMarketplace.capabilities.tiered") }}
                  </span>
                  <span v-if="model.capabilities.has_per_request_pricing">
                    {{ t("modelMarketplace.capabilities.perRequest") }}
                  </span>
                  <span
                    v-if="!hasCapabilities(model)"
                    class="text-muted-soft"
                  >
                    -
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
            <span>{{ t("modelMarketplace.columns.channels") }}</span>
            <strong>{{ selectedModel.channel_count }}</strong>
          </div>
          <div class="detail-metric">
            <span>{{ t("modelMarketplace.columns.groups") }}</span>
            <strong>{{ selectedModel.group_count }}</strong>
          </div>
          <div class="detail-metric">
            <span>{{ t("modelMarketplace.columns.priceVariants") }}</span>
            <strong>{{ selectedModel.price_variant_count }}</strong>
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

        <div class="channel-detail-list">
          <article
            v-for="channel in selectedModel.channels"
            :key="`${channel.id}:${channel.platform}`"
            class="channel-detail"
          >
            <div class="channel-detail-header">
              <div>
                <h3>{{ channel.name }}</h3>
                <p v-if="channel.description">{{ channel.description }}</p>
              </div>
              <span>{{ pricingSourceLabel(channel.pricing_source) }}</span>
            </div>

            <div class="mapping-line">
              <span>{{ t("modelMarketplace.mapping") }}</span>
              <code>{{ channel.mapping.chain }}</code>
            </div>

            <div class="pricing-grid">
              <div>
                <span>{{ t("modelMarketplace.pricing.input") }}</span>
                <strong>{{ formatMillion(channel.pricing?.input_price ?? null) }}</strong>
              </div>
              <div>
                <span>{{ t("modelMarketplace.pricing.output") }}</span>
                <strong>{{ formatMillion(channel.pricing?.output_price ?? null) }}</strong>
              </div>
              <div>
                <span>{{ t("modelMarketplace.pricing.cacheWrite") }}</span>
                <strong>{{ formatMillion(channel.pricing?.cache_write_price ?? null) }}</strong>
              </div>
              <div>
                <span>{{ t("modelMarketplace.pricing.cacheRead") }}</span>
                <strong>{{ formatMillion(channel.pricing?.cache_read_price ?? null) }}</strong>
              </div>
            </div>

            <div class="group-list">
              <div
                v-for="group in channel.groups"
                :key="group.id"
                class="group-row"
              >
                <div>
                  <span class="group-name">{{ group.name }}</span>
                  <span class="group-kind">
                    {{
                      group.is_exclusive
                        ? t("modelMarketplace.groups.exclusive")
                        : t("modelMarketplace.groups.public")
                    }}
                  </span>
                </div>
                <div class="group-rate">
                  <span>
                    {{
                      t("modelMarketplace.groups.rate", {
                        rate: formatMultiplier(group.rate_multiplier),
                      })
                    }}
                  </span>
                  <strong v-if="group.user_rate_multiplier != null">
                    {{
                      t("modelMarketplace.groups.userRate", {
                        rate: formatMultiplier(group.user_rate_multiplier),
                      })
                    }}
                  </strong>
                  <strong v-else>
                    {{
                      t("modelMarketplace.groups.effectiveRate", {
                        rate: formatMultiplier(group.effective_rate_multiplier),
                      })
                    }}
                  </strong>
                </div>
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

type SortKey = "name" | "platform" | "input" | "output" | "channels";
type ViewMode = "table" | "cards";

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
  channel: "all",
  group: "all",
  billing: "all",
  pricing: "all",
  capability: "all",
});

const models = computed(() => response.value.models);
const summary = computed(() => response.value.summary);

const summaryCards = computed(() => [
  {
    key: "models",
    label: t("modelMarketplace.summary.models"),
    value: summary.value.models,
    icon: "cube" as const,
  },
  {
    key: "platforms",
    label: t("modelMarketplace.summary.platforms"),
    value: summary.value.platforms,
    icon: "grid" as const,
  },
  {
    key: "channels",
    label: t("modelMarketplace.summary.channels"),
    value: summary.value.channels,
    icon: "server" as const,
  },
  {
    key: "priced",
    label: t("modelMarketplace.summary.priced"),
    value: summary.value.priced_models,
    icon: "dollar" as const,
  },
]);

const platformOptions = computed(() =>
  Array.from(new Set(models.value.map((m) => m.platform))).sort(),
);

const channelOptions = computed(() =>
  Array.from(
    new Set(models.value.flatMap((m) => m.channels.map((c) => c.name))),
  ).sort((a, b) => a.localeCompare(b)),
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

const filteredModels = computed(() => {
  const q = filters.search.trim().toLowerCase();
  const rows = models.value.filter((model) => {
    if (filters.platform !== "all" && model.platform !== filters.platform) {
      return false;
    }
    if (filters.billing !== "all" && model.billing_mode !== filters.billing) {
      return false;
    }
    if (filters.pricing === "priced" && !model.pricing) return false;
    if (filters.pricing === "unpriced" && model.pricing) return false;
    if (!matchesCapability(model, filters.capability)) return false;
    if (
      filters.channel !== "all" &&
      !model.channels.some((c) => c.name === filters.channel)
    ) {
      return false;
    }
    if (
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
      model.channels.some(
        (channel) =>
          channel.name.toLowerCase().includes(q) ||
          channel.groups.some((g) => g.name.toLowerCase().includes(q)),
      )
    );
  });

  return [...rows].sort((a, b) => compareModels(a, b, sortKey.value));
});

function matchesCapability(model: ModelMarketplaceModel, capability: string) {
  switch (capability) {
    case "image":
      return model.capabilities.supports_image;
    case "cache":
      return model.capabilities.supports_cache_pricing;
    case "tiered":
      return model.capabilities.has_tiered_pricing;
    case "per_request":
      return model.capabilities.has_per_request_pricing;
    default:
      return true;
  }
}

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
    case "channels":
      return b.channel_count - a.channel_count || a.id.localeCompare(b.id);
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

function resetFilters() {
  filters.search = "";
  filters.platform = "all";
  filters.channel = "all";
  filters.group = "all";
  filters.billing = "all";
  filters.pricing = "all";
  filters.capability = "all";
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

function hasCapabilities(model: ModelMarketplaceModel) {
  return (
    model.capabilities.supports_image ||
    model.capabilities.supports_cache_pricing ||
    model.capabilities.has_tiered_pricing ||
    model.capabilities.has_per_request_pricing
  );
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

.summary-grid {
  @apply grid gap-3 sm:grid-cols-2 xl:grid-cols-4;
}

.summary-card {
  @apply flex items-center gap-4 rounded-lg border border-hairline bg-canvas p-5;
}

.summary-icon {
  @apply flex h-10 w-10 items-center justify-center rounded-md bg-surface-card text-primary-700;
}

.summary-label {
  @apply text-xs font-medium uppercase tracking-wide text-muted;
}

.summary-value {
  @apply mt-1 font-serif text-3xl font-normal leading-none text-ink;
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
  @apply grid gap-3 md:grid-cols-2 xl:grid-cols-6;
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
  background-position: right 0.85rem center;
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

.model-badges,
.capability-list {
  @apply flex flex-wrap gap-1.5;
}

.model-badges span,
.capability-list span,
.platform-pill {
  @apply inline-flex items-center gap-1 rounded-full bg-surface-card px-2.5 py-1 text-xs font-medium text-body;
}

.model-card-actions {
  @apply mt-auto flex justify-end;
}

.model-table-wrap {
  @apply overflow-x-auto rounded-lg border border-hairline bg-canvas;
}

.model-table {
  @apply w-full min-w-[980px] border-collapse text-sm;
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
.price-cell {
  @apply flex flex-col gap-1;
}

.model-id-button {
  @apply break-all text-left font-medium text-ink hover:text-primary-700;
}

.model-cell span,
.price-cell span {
  @apply text-xs text-muted;
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
  @apply grid gap-3 md:grid-cols-4;
}

.detail-metric {
  @apply rounded-lg border border-hairline bg-canvas p-4;
}

.detail-metric span {
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

.channel-detail-list {
  @apply space-y-4;
}

.channel-detail {
  @apply rounded-lg border border-hairline bg-canvas p-5;
}

.channel-detail-header {
  @apply flex flex-col justify-between gap-3 md:flex-row md:items-start;
}

.channel-detail-header h3 {
  @apply text-lg font-medium text-ink;
}

.channel-detail-header p {
  @apply mt-1 text-sm text-muted;
}

.channel-detail-header > span {
  @apply inline-flex w-fit rounded-full bg-surface-card px-3 py-1 text-xs font-medium text-body;
}

.mapping-line {
  @apply mt-4 flex flex-col gap-2 rounded-md bg-surface-soft p-3 text-sm md:flex-row md:items-center;
}

.mapping-line span {
  @apply text-xs font-medium uppercase tracking-wide text-muted;
}

.mapping-line code {
  @apply break-all font-mono text-sm text-ink;
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

.group-list {
  @apply mt-4 divide-y divide-hairline rounded-md border border-hairline;
}

.group-row {
  @apply flex flex-col justify-between gap-3 px-3 py-3 text-sm md:flex-row md:items-center;
}

.group-name {
  @apply font-medium text-ink;
}

.group-kind {
  @apply ml-2 rounded-full bg-surface-card px-2 py-0.5 text-xs text-muted;
}

.group-rate {
  @apply flex flex-wrap items-center gap-2 text-xs text-muted;
}

.group-rate strong {
  @apply rounded-full bg-primary-100 px-2 py-0.5 font-medium text-primary-800;
}
</style>
