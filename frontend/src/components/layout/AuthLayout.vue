<template>
  <div
    class="relative flex min-h-screen items-center justify-center overflow-hidden bg-canvas p-4"
  >
    <div
      class="absolute inset-x-0 top-0 h-24 border-b border-hairline bg-cream/65"
    ></div>
    <div
      class="relative z-10 grid w-full max-w-5xl overflow-hidden rounded-lg border border-hairline bg-canvas shadow-card-hover lg:grid-cols-[1fr_440px]"
    >
      <section
        class="hidden min-h-[620px] flex-col justify-between bg-surface-dark p-10 text-canvas lg:flex"
      >
        <div>
          <div class="mb-10 flex items-center gap-3">
            <div
              class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-lg bg-canvas"
            >
              <img
                :src="siteLogo || '/logo.png'"
                alt="Logo"
                class="h-full w-full object-contain"
              />
            </div>
            <span class="font-display text-2xl font-medium">{{
              siteName
            }}</span>
          </div>
          <p
            class="mb-5 text-xs font-medium uppercase tracking-[0.22em] text-primary-200"
          >
            API Gateway
          </p>
          <h1
            class="font-display text-5xl font-medium leading-[1.04] text-canvas"
          >
            {{ siteSubtitle }}
          </h1>
        </div>
        <div
          class="rounded-lg border border-on-dark/10 bg-surface-dark-elevated p-5 font-mono text-sm text-canvas shadow-card"
        >
          <div class="mb-4 flex items-center gap-2 text-muted-soft">
            <span class="h-2.5 w-2.5 rounded-full bg-primary-500"></span>
            <span class="h-2.5 w-2.5 rounded-full bg-accent-amber"></span>
            <span class="h-2.5 w-2.5 rounded-full bg-success"></span>
            <span class="ml-auto text-xs">sub2api</span>
          </div>
          <div class="space-y-2">
            <p><span class="text-primary-300">$</span> route /v1/messages</p>
            <p class="text-muted-soft">account pool resolved</p>
            <p>
              <span class="text-success">200 OK</span>
              <span class="text-primary-200">usage tracked</span>
            </p>
          </div>
        </div>
      </section>

      <section class="px-6 py-8 sm:px-10">
        <div class="mb-8 text-center lg:hidden">
          <template v-if="settingsLoaded">
            <div
              class="mb-4 inline-flex h-14 w-14 items-center justify-center overflow-hidden rounded-lg border border-hairline bg-canvas shadow-sm"
            >
              <img
                :src="siteLogo || '/logo.png'"
                alt="Logo"
                class="h-full w-full object-contain"
              />
            </div>
            <h1 class="mb-2 font-display text-4xl font-medium text-ink">
              {{ siteName }}
            </h1>
            <p class="text-sm text-muted">
              {{ siteSubtitle }}
            </p>
          </template>
        </div>

        <div class="rounded-lg border border-hairline bg-cream/45 p-6 sm:p-8">
          <slot />
        </div>

        <div class="mt-6 text-center text-sm text-muted">
          <slot name="footer" />
        </div>

        <div class="mt-8 text-center text-xs text-muted-soft">
          &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useAppStore } from "@/stores";
import { sanitizeUrl } from "@/utils/url";

const appStore = useAppStore();

const siteName = computed(() => appStore.siteName || "Sub2API");
const siteLogo = computed(() =>
  sanitizeUrl(appStore.siteLogo || "", {
    allowRelative: true,
    allowDataUrl: true,
  }),
);
const siteSubtitle = computed(
  () =>
    appStore.cachedPublicSettings?.site_subtitle ||
    "Subscription to API Conversion Platform",
);
const settingsLoaded = computed(() => appStore.publicSettingsLoaded);

const currentYear = computed(() => new Date().getFullYear());

onMounted(() => {
  appStore.fetchPublicSettings();
});
</script>
