<template>
  <AppLayout>
    <div class="space-y-4">
      <div
        class="grid gap-3 rounded-lg border border-hairline bg-canvas p-4 lg:grid-cols-[minmax(0,1fr)_18rem]"
      >
        <div class="grid gap-3 md:grid-cols-2">
          <label class="block">
            <span class="input-label">{{ t("playground.apiKey") }}</span>
            <select v-model="selectedKeyId" class="input">
              <option value="manual">{{ t("playground.manualKey") }}</option>
              <option v-for="key in apiKeys" :key="key.id" :value="String(key.id)">
                {{ key.name }} · {{ key.group?.name || t("common.default") }}
              </option>
            </select>
          </label>
          <label class="block">
            <span class="input-label">{{ t("playground.model") }}</span>
            <input v-model="chatModel" class="input" list="playground-models" />
          </label>
        </div>
        <label class="block">
          <span class="input-label">{{ t("playground.manualKey") }}</span>
          <input
            v-model="manualKey"
            class="input"
            autocomplete="off"
            :placeholder="t('playground.manualKeyPlaceholder')"
            type="password"
          />
        </label>
      </div>

      <div class="flex flex-wrap gap-2 border-b border-hairline pb-2">
        <button
          class="rounded-lg px-3 py-2 text-sm font-medium transition-colors"
          :class="mode === 'chat' ? activeTabClass : idleTabClass"
          @click="mode = 'chat'"
        >
          <Icon name="chat" size="sm" class="mr-1 inline" />
          {{ t("playground.chat") }}
        </button>
        <button
          class="rounded-lg px-3 py-2 text-sm font-medium transition-colors"
          :class="mode === 'image' ? activeTabClass : idleTabClass"
          @click="mode = 'image'"
        >
          <Icon name="sparkles" size="sm" class="mr-1 inline" />
          {{ t("playground.image") }}
        </button>
      </div>

      <section v-if="mode === 'chat'" class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_18rem]">
        <div class="card flex min-h-[32rem] flex-col p-0">
          <div class="flex-1 space-y-3 overflow-y-auto p-4">
            <div
              v-for="message in messages"
              :key="message.id"
              class="flex"
              :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
            >
              <div
                class="max-w-[85%] whitespace-pre-wrap rounded-lg px-3 py-2 text-sm"
                :class="
                  message.role === 'user'
                    ? 'bg-primary-600 text-on-primary'
                    : 'bg-surface-card text-body'
                "
              >
                {{ message.content }}
              </div>
            </div>
            <div v-if="chatLoading" class="text-sm text-muted">
              {{ t("common.loading") }}
            </div>
          </div>
          <form
            class="border-t border-hairline p-3"
            @submit.prevent="sendChatMessage"
          >
            <textarea
              v-model="chatInput"
              class="input min-h-24 resize-y"
              :placeholder="t('playground.messagePlaceholder')"
              @keydown.enter.exact.prevent="sendChatMessage"
            ></textarea>
            <div class="mt-2 flex justify-end">
              <button class="btn btn-primary" :disabled="chatLoading">
                <Icon name="play" size="sm" />
                {{ t("playground.send") }}
              </button>
            </div>
          </form>
        </div>
        <label class="block">
          <span class="input-label">{{ t("playground.systemPrompt") }}</span>
          <textarea v-model="systemPrompt" class="input min-h-48 resize-y"></textarea>
        </label>
      </section>

      <section v-else class="grid gap-4 lg:grid-cols-[20rem_minmax(0,1fr)]">
        <form class="card space-y-3 p-4" @submit.prevent="generateImage">
          <label class="block">
            <span class="input-label">{{ t("playground.imageModel") }}</span>
            <input v-model="imageModel" class="input" list="playground-models" />
          </label>
          <label class="block">
            <span class="input-label">{{ t("playground.imageSize") }}</span>
            <select v-model="imageSize" class="input">
              <option>1024x1024</option>
              <option>1024x1536</option>
              <option>1536x1024</option>
            </select>
          </label>
          <label class="block">
            <span class="input-label">{{ t("playground.prompt") }}</span>
            <textarea
              v-model="imagePrompt"
              class="input min-h-48 resize-y"
              :placeholder="t('playground.imagePromptPlaceholder')"
            ></textarea>
          </label>
          <button class="btn btn-primary w-full" :disabled="imageLoading">
            <Icon name="sparkles" size="sm" />
            {{ imageLoading ? t("common.loading") : t("playground.generate") }}
          </button>
        </form>
        <div class="card flex min-h-[32rem] items-center justify-center p-4">
          <img
            v-if="imageUrl"
            :src="imageUrl"
            class="max-h-[70vh] max-w-full rounded-lg object-contain"
            alt=""
          />
          <div v-else class="text-sm text-muted">
            {{ t("playground.noImage") }}
          </div>
        </div>
      </section>

      <datalist id="playground-models">
        <option v-for="model in modelOptions" :key="model" :value="model" />
      </datalist>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import keysAPI from "@/api/keys";
import Icon from "@/components/icons/Icon.vue";
import AppLayout from "@/components/layout/AppLayout.vue";
import { useAppStore } from "@/stores/app";
import type { ApiKey } from "@/types";
import { allModels } from "@/composables/useModelWhitelist";

type PlaygroundMode = "chat" | "image";
type ChatMessage = {
  id: string;
  role: "user" | "assistant";
  content: string;
};

const { t } = useI18n();
const appStore = useAppStore();
const apiKeys = ref<ApiKey[]>([]);
const selectedKeyId = ref(localStorage.getItem("playground.keyId") || "manual");
localStorage.removeItem("playground.manualKey");
const manualKey = ref("");
const mode = ref<PlaygroundMode>("chat");
const chatModel = ref(localStorage.getItem("playground.chatModel") || "gpt-4o");
const imageModel = ref(
  localStorage.getItem("playground.imageModel") || "gpt-image-2",
);
const imageSize = ref(localStorage.getItem("playground.imageSize") || "1024x1024");
const systemPrompt = ref("");
const chatInput = ref("");
const imagePrompt = ref("");
const messages = ref<ChatMessage[]>([]);
const chatLoading = ref(false);
const imageLoading = ref(false);
const imageUrl = ref("");
const activeTabClass = "bg-primary-100 text-primary-700";
const idleTabClass = "text-muted hover:bg-surface-card hover:text-body";
const modelOptions = computed(() => [
  ...new Set([...allModels.map((model) => model.value), "gpt-image-2"]),
]);

const selectedKey = computed(() =>
  apiKeys.value.find((key) => String(key.id) === selectedKeyId.value),
);
const resolvedKey = computed(() =>
  selectedKeyId.value === "manual" ? manualKey.value.trim() : selectedKey.value?.key || "",
);

watch(selectedKeyId, (value) => localStorage.setItem("playground.keyId", value));
watch(chatModel, (value) => localStorage.setItem("playground.chatModel", value));
watch(imageModel, (value) => localStorage.setItem("playground.imageModel", value));
watch(imageSize, (value) => localStorage.setItem("playground.imageSize", value));

async function loadKeys() {
  try {
    const result = await keysAPI.list(1, 100, { status: "active" });
    apiKeys.value = result.items;
    if (selectedKeyId.value !== "manual" && !selectedKey.value) {
      selectedKeyId.value = result.items[0] ? String(result.items[0].id) : "manual";
    }
  } catch (error) {
    console.error("Failed to load API keys: ", error);
  }
}

function ensureKey(): string | null {
  const key = resolvedKey.value;
  if (!key) {
    appStore.showError(t("playground.keyRequired"));
    return null;
  }
  return key;
}

async function postOpenAICompatible<T>(path: string, payload: unknown): Promise<T> {
  const key = ensureKey();
  if (!key) throw new Error("Missing API key");
  const response = await fetch(path, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${key}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    const message =
      data?.error?.message || data?.message || `${response.status} ${response.statusText}`;
    throw new Error(message);
  }
  return data as T;
}

async function sendChatMessage() {
  const content = chatInput.value.trim();
  if (!content || chatLoading.value) return;
  chatLoading.value = true;
  const userMessage: ChatMessage = {
    id: `u-${Date.now()}`,
    role: "user",
    content,
  };
  messages.value.push(userMessage);
  chatInput.value = "";
  try {
    const payloadMessages = [
      ...(systemPrompt.value.trim()
        ? [{ role: "system", content: systemPrompt.value.trim() }]
        : []),
      ...messages.value.map(({ role, content }) => ({ role, content })),
    ];
    const data = await postOpenAICompatible<{
      choices?: Array<{ message?: { content?: string } }>;
    }>("/v1/chat/completions", {
      model: chatModel.value,
      messages: payloadMessages,
    });
    messages.value.push({
      id: `a-${Date.now()}`,
      role: "assistant",
      content: data.choices?.[0]?.message?.content || "",
    });
  } catch (error: any) {
    appStore.showError(error?.message || t("common.error"));
    messages.value = messages.value.filter((message) => message.id !== userMessage.id);
  } finally {
    chatLoading.value = false;
  }
}

async function generateImage() {
  const prompt = imagePrompt.value.trim();
  if (!prompt || imageLoading.value) return;
  imageLoading.value = true;
  try {
    const data = await postOpenAICompatible<{
      data?: Array<{ url?: string; b64_json?: string }>;
    }>("/v1/images/generations", {
      model: imageModel.value,
      prompt,
      size: imageSize.value,
      n: 1,
    });
    const image = data.data?.[0];
    imageUrl.value = image?.url || (image?.b64_json ? `data:image/png;base64,${image.b64_json}` : "");
  } catch (error: any) {
    appStore.showError(error?.message || t("common.error"));
  } finally {
    imageLoading.value = false;
  }
}

onMounted(loadKeys);
</script>
