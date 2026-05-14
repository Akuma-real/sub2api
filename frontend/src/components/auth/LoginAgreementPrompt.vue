<template>
  <div v-if="mode === 'checkbox' && documents.length > 0" class="px-0.5">
    <div class="flex items-start gap-2">
      <input
        id="login-agreement-consent"
        type="checkbox"
        :checked="accepted"
        class="mt-[2px] h-4 w-4 flex-shrink-0 rounded border-hairline text-primary-600 focus:ring-primary-500"
        @change="handleCheckboxChange"
      />
      <div class="min-w-0 flex-1">
        <p class="text-[13px] leading-5 text-body">
          <label for="login-agreement-consent" class="cursor-pointer text-body">
            我已阅读并同意
          </label>
          <template
            v-for="(doc, index) in documents"
            :key="doc.id || doc.title"
          >
            <RouterLink
              :to="documentRoute(doc)"
              target="_blank"
              rel="noopener noreferrer"
              class="font-medium text-primary-600 underline-offset-4 transition hover:text-primary-700 hover:underline"
            >
              {{ doc.title }}
            </RouterLink>
            <span v-if="index < documents.length - 1">、</span>
          </template>
        </p>
      </div>
    </div>
  </div>

  <div
    v-else-if="!accepted && documents.length > 0"
    class="rounded-lg border border-primary-100 bg-primary-50/70 p-3 text-sm text-primary-900"
  >
    <div class="flex items-start gap-3">
      <Icon
        name="shield"
        size="sm"
        class="mt-0.5 flex-shrink-0 text-primary-600"
      />
      <div class="min-w-0 flex-1">
        <p class="font-medium">继续登录前需要先同意最新条款。</p>
        <p class="mt-1 text-primary-700">
          未同意前，账号密码输入和快捷登录会保持禁用。
        </p>
      </div>
      <button
        type="button"
        class="flex-shrink-0 rounded-md bg-primary-600 px-3 py-1.5 text-xs font-medium text-on-primary transition hover:bg-primary-700"
        @click="emit('open')"
      >
        查看条款
      </button>
    </div>
  </div>

  <Teleport to="body">
    <Transition name="agreement-fade">
      <div
        v-if="dialogVisible"
        class="fixed inset-0 z-[140] flex items-center justify-center overflow-y-auto bg-surface-dark/60 p-4 "
      >
        <div
          class="w-full max-w-[600px] overflow-hidden rounded-lg bg-canvas shadow-card ring-1 ring-ink/10"
        >
          <div class="border-b border-hairline-soft bg-canvas px-6 py-6">
            <div class="flex items-start gap-4">
              <span
                class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-primary-50 text-primary-700 ring-1 ring-primary-100"
              >
                <Icon name="shield" size="md" />
              </span>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <h2 class="text-xl font-bold tracking-normal text-ink">
                    条款更新通知
                  </h2>
                  <span
                    v-if="updatedAt"
                    class="rounded-full bg-surface-card px-2.5 py-1 text-xs font-medium text-body"
                  >
                    {{ updatedAt }}
                  </span>
                </div>
                <p class="mt-2 text-sm leading-6 text-body">
                  我们的服务条款已于
                  {{
                    updatedAt || "近期"
                  }}
                  更新。在继续使用服务之前，请仔细阅读并同意以下条款。
                </p>
              </div>
            </div>
          </div>

          <div class="max-h-[58vh] overflow-y-auto px-6 py-5">
            <div class="mb-3 flex items-center justify-between gap-3">
              <p class="text-sm font-semibold text-ink">相关文档</p>
            </div>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <RouterLink
                v-for="(doc, index) in documents"
                :key="doc.id || doc.title"
                :to="documentRoute(doc)"
                target="_blank"
                rel="noopener noreferrer"
                class="group flex min-h-[72px] w-full items-center gap-3 rounded-xl border border-hairline bg-surface-soft/70 px-4 py-3 text-left transition hover:-translate-y-0.5 hover:border-primary-200 hover:bg-canvas hover:shadow-sm"
              >
                <span
                  class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-canvas text-body ring-1 ring-hairline transition group-hover:bg-primary-50 group-hover:text-primary-700 group-hover:ring-primary-100"
                >
                  <Icon :name="documentIcon(index, doc.title)" size="sm" />
                </span>
                <span class="min-w-0 flex-1">
                  <span class="block truncate text-sm font-semibold text-ink">{{
                    doc.title
                  }}</span>
                </span>
                <span
                  class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-muted-soft transition group-hover:bg-primary-50 group-hover:text-primary-600"
                >
                  <Icon name="externalLink" size="sm" />
                </span>
              </RouterLink>
            </div>
          </div>

          <div
            class="border-t border-hairline-soft bg-surface-soft/80 px-6 py-4"
          >
            <div class="grid grid-cols-2 gap-3">
              <button
                type="button"
                class="rounded-xl border border-hairline bg-canvas px-4 py-3 text-sm font-semibold text-body transition hover:bg-surface-card"
                @click="emit('reject')"
              >
                拒绝
              </button>
              <button
                type="button"
                class="rounded-xl bg-primary-600 px-4 py-3 text-sm font-semibold text-on-primary shadow-sm shadow-primary-600/20 transition hover:bg-primary-700"
                @click="emit('accept')"
              >
                同意并继续
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from "vue";
import Icon from "@/components/icons/Icon.vue";
import type { LoginAgreementDocument } from "@/types";

const props = withDefaults(
  defineProps<{
    accepted: boolean;
    documents: LoginAgreementDocument[];
    mode: "modal" | "checkbox" | string;
    updatedAt?: string;
    visible: boolean;
  }>(),
  {
    updatedAt: "",
  },
);

const emit = defineEmits<{
  accept: [];
  reject: [];
  open: [];
}>();

const dialogVisible = computed(
  () => props.visible && documents.value.length > 0,
);
const documents = computed(() =>
  props.documents.filter((doc) => doc.title.trim()),
);
const updatedAt = computed(() => props.updatedAt || "");
const accepted = computed(() => props.accepted);
const mode = computed(() => (props.mode === "checkbox" ? "checkbox" : "modal"));

function documentRoute(doc: LoginAgreementDocument) {
  return {
    name: "LegalDocument",
    params: {
      documentId: doc.id || doc.title,
    },
  };
}

function handleCheckboxChange(event: Event): void {
  const checked = (event.target as HTMLInputElement).checked;
  if (checked) {
    emit("accept");
  } else {
    emit("reject");
  }
}

function documentIcon(
  index: number,
  title: string,
): "document" | "shield" | "globe" | "cog" {
  if (title.includes("政策") || title.includes("隐私")) {
    return "shield";
  }
  if (title.includes("国家") || title.includes("地区")) {
    return "globe";
  }
  if (index === 3) {
    return "cog";
  }
  return "document";
}
</script>

<style scoped>
.agreement-fade-enter-active,
.agreement-fade-leave-active {
  transition: opacity 0.18s ease;
}

.agreement-fade-enter-from,
.agreement-fade-leave-to {
  opacity: 0;
}

.agreement-fade-enter-active > div,
.agreement-fade-leave-active > div {
  transition:
    transform 0.18s ease,
    opacity 0.18s ease;
}

.agreement-fade-enter-from > div,
.agreement-fade-leave-to > div {
  opacity: 0;
  transform: translateY(8px) scale(0.98);
}
</style>
