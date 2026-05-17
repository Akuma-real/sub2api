<template>
  <div>
    <!-- Tags display -->
    <div class="mb-1 flex justify-end" v-if="models.length > 0">
      <button
        type="button"
        class="inline-flex items-center gap-1 text-xs text-primary-600 hover:text-primary-700"
        @click="copyModels"
      >
        <Icon name="copy" size="xs" />
        {{ t("common.copy") }}
      </button>
    </div>
    <div class="relative">
      <div
        class="flex min-h-[2.5rem] flex-wrap gap-1.5 rounded-lg border border-hairline bg-canvas p-2"
      >
      <span
        v-for="(model, idx) in models"
        :key="idx"
        class="inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-sm"
        :class="getPlatformTagClass(props.platform || '')"
      >
        {{ model }}
        <button
          type="button"
          @click="removeModel(idx)"
          class="ml-0.5 rounded-full p-0.5 hover:bg-primary-200"
        >
          <Icon name="x" size="xs" />
        </button>
      </span>
      <input
        ref="inputRef"
        v-model="inputValue"
        type="text"
        class="min-w-[120px] flex-1 border-none bg-transparent text-sm outline-none placeholder:text-muted-soft"
        :placeholder="models.length === 0 ? placeholder : ''"
        @keydown.enter.prevent="addModel"
        @keydown.tab.prevent="addModel"
        @keydown.delete="handleBackspace"
        @paste="handlePaste"
        @blur="addModel"
        @focus="showSuggestions = true"
      />
      </div>
      <div
        v-if="showSuggestions && suggestions.length > 0"
        class="absolute z-30 mt-1 max-h-56 w-full overflow-y-auto rounded-lg border border-hairline bg-canvas p-1 shadow-card"
      >
        <button
          v-for="model in suggestions"
          :key="model"
          type="button"
          class="flex w-full items-center justify-between rounded-md px-2 py-1.5 text-left text-sm text-body hover:bg-surface-card"
          @mousedown.prevent="selectSuggestion(model)"
        >
          <span class="truncate">{{ model }}</span>
          <Icon name="plus" size="xs" class="text-muted-soft" />
        </button>
      </div>
    </div>
    <p class="mt-1 text-xs text-muted-soft">
      {{
        t(
          "admin.channels.form.modelInputHint",
          "Press Enter to add, supports paste for batch import.",
        )
      }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";
import { allModels, getModelsByPlatform } from "@/composables/useModelWhitelist";
import { getPlatformTagClass } from "./types";

const { t } = useI18n();

const props = defineProps<{
  models: string[];
  placeholder?: string;
  platform?: string;
}>();

const emit = defineEmits<{
  "update:models": [models: string[]];
}>();

const inputValue = ref("");
const inputRef = ref<HTMLInputElement>();
const showSuggestions = ref(false);

const suggestions = computed(() => {
  const keyword = inputValue.value.trim().toLowerCase();
  const source =
    props.platform && getModelsByPlatform(props.platform).length > 0
      ? getModelsByPlatform(props.platform)
      : allModels.map((model) => model.value);
  return source
    .filter((model) => !props.models.includes(model))
    .filter((model) => !keyword || model.toLowerCase().includes(keyword))
    .slice(0, 8);
});

function addModel() {
  const val = inputValue.value.trim();
  if (!val) return;
  if (!props.models.includes(val)) {
    emit("update:models", [...props.models, val]);
  }
  inputValue.value = "";
  showSuggestions.value = false;
}

function selectSuggestion(model: string) {
  if (!props.models.includes(model)) {
    emit("update:models", [...props.models, model]);
  }
  inputValue.value = "";
  showSuggestions.value = false;
}

function removeModel(idx: number) {
  const newModels = [...props.models];
  newModels.splice(idx, 1);
  emit("update:models", newModels);
}

function handleBackspace() {
  if (inputValue.value === "" && props.models.length > 0) {
    removeModel(props.models.length - 1);
  }
}

function handlePaste(e: ClipboardEvent) {
  e.preventDefault();
  const text = e.clipboardData?.getData("text") || "";
  const items = text
    .split(/[,\n;]+/)
    .map((s) => s.trim())
    .filter(Boolean);
  if (items.length === 0) return;
  const unique = [...new Set([...props.models, ...items])];
  emit("update:models", unique);
  inputValue.value = "";
}

async function copyModels() {
  await navigator.clipboard.writeText(props.models.join(","));
}
</script>
