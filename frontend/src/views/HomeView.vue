<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div
    v-else
    class="relative flex min-h-screen flex-col overflow-hidden bg-canvas text-ink"
  >
    <header
      class="relative z-20 border-b border-hairline bg-canvas/90 px-6 py-4 "
    >
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <div class="flex items-center gap-3">
          <div
            class="h-10 w-10 overflow-hidden rounded-lg border border-hairline bg-canvas shadow-sm"
          >
            <img
              :src="siteLogo || '/logo.png'"
              alt="Logo"
              class="h-full w-full object-contain"
            />
          </div>
          <span
            class="hidden font-display text-2xl font-medium leading-none text-ink sm:inline"
          >
            {{ siteName }}
          </span>
        </div>

        <div class="flex items-center gap-3">
          <LocaleSwitcher />

          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="rounded-lg p-2 text-muted transition-colors hover:bg-cream hover:text-ink"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center gap-1.5 rounded-full bg-surface-dark py-1 pl-1 pr-2.5 transition-colors hover:bg-surface-dark-elevated"
          >
            <span
              class="flex h-5 w-5 items-center justify-center rounded-full bg-primary-500 text-[10px] font-semibold text-on-primary"
            >
              {{ userInitial }}
            </span>
            <span class="text-xs font-medium text-on-primary">{{
              t("home.dashboard")
            }}</span>
            <svg
              class="h-3 w-3 text-muted-soft"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25"
              />
            </svg>
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center rounded-full bg-surface-dark px-4 py-2 text-xs font-medium text-on-dark transition-colors hover:bg-surface-dark-elevated"
          >
            {{ t("home.login") }}
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="relative z-10 flex-1 px-6">
      <div class="mx-auto max-w-6xl">
        <section
          class="grid min-h-[calc(100vh-5rem)] items-center gap-10 py-14 lg:grid-cols-[0.95fr_1.05fr] lg:py-20"
        >
          <div class="text-center lg:text-left">
            <p
              class="mb-5 text-xs font-medium uppercase tracking-[0.22em] text-primary-700"
            >
              API Gateway Platform
            </p>
            <h1
              class="mb-5 font-display text-5xl font-medium leading-[1.02] text-ink md:text-6xl lg:text-7xl"
            >
              {{ siteName }}
            </h1>
            <p
              class="mx-auto mb-8 max-w-2xl text-lg leading-8 text-body lg:mx-0 md:text-xl"
            >
              {{ siteSubtitle }}
            </p>

            <div
              class="flex flex-col items-center gap-3 sm:flex-row lg:items-start"
            >
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="btn btn-primary px-8 py-3 text-base"
              >
                {{
                  isAuthenticated
                    ? t("home.goToDashboard")
                    : t("home.getStarted")
                }}
                <Icon
                  name="arrowRight"
                  size="md"
                  class="ml-2"
                  :stroke-width="2"
                />
              </router-link>
              <a
                v-if="docUrl"
                :href="docUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-secondary px-8 py-3 text-base"
              >
                {{ t("home.viewDocs") }}
              </a>
            </div>
          </div>

          <div class="flex justify-center lg:justify-end">
            <div class="terminal-container">
              <div class="terminal-window">
                <!-- Window header -->
                <div class="terminal-header">
                  <div class="terminal-buttons">
                    <span class="btn-close"></span>
                    <span class="btn-minimize"></span>
                    <span class="btn-maximize"></span>
                  </div>
                  <span class="terminal-title">terminal</span>
                </div>
                <!-- Terminal content -->
                <div class="terminal-body">
                  <div class="code-line line-1">
                    <span class="code-prompt">{{ shellUsername }}@chatgpt</span>
                    <span class="code-path">~ $</span>
                    <span class="code-cmd">motd</span>
                  </div>
                  <div class="code-line line-2">
                    <span class="code-welcome">Welcome to {{ siteName }}</span>
                  </div>
                  <div class="code-line line-3">
                    <span class="code-success">认证入口已在线</span>
                  </div>
                  <template v-if="geoStatus === 'loading'">
                    <div class="code-line line-4">
                      <span class="code-comment">正在解析地理位置...</span>
                    </div>
                  </template>
                  <template v-else-if="geoStatus === 'success'">
                    <div class="code-line motd-row line-4">
                      <span class="code-label">IP</span>
                      <span class="code-value">{{ geoInfo.ip }}</span>
                    </div>
                    <div class="code-line motd-row line-5">
                      <span class="code-label">ASN</span>
                      <span class="code-value code-value-wrap">{{
                        geoInfo.asnText
                      }}</span>
                    </div>
                    <div class="code-line motd-row line-6">
                      <span class="code-label">来源国家</span>
                      <span class="code-value">{{ geoInfo.countryName }}</span>
                    </div>
                  </template>
                  <template v-else>
                    <div class="code-line line-4">
                      <span class="code-error">IP 解析失败</span>
                    </div>
                  </template>
                  <div class="code-line line-cursor">
                    <span class="code-prompt">$</span>
                    <span class="cursor"></span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <div
          class="mb-12 flex flex-wrap items-center justify-center gap-3 md:gap-4"
        >
          <div
            class="inline-flex items-center gap-2.5 rounded-full border border-hairline bg-cream px-5 py-2.5 shadow-sm"
          >
            <Icon name="swap" size="sm" class="text-primary-700" />
            <span class="text-sm font-medium text-ink">{{
              t("home.tags.subscriptionToApi")
            }}</span>
          </div>
          <div
            class="inline-flex items-center gap-2.5 rounded-full border border-hairline bg-cream px-5 py-2.5 shadow-sm"
          >
            <Icon name="shield" size="sm" class="text-primary-700" />
            <span class="text-sm font-medium text-ink">{{
              t("home.tags.stickySession")
            }}</span>
          </div>
          <div
            class="inline-flex items-center gap-2.5 rounded-full border border-hairline bg-cream px-5 py-2.5 shadow-sm"
          >
            <Icon name="chart" size="sm" class="text-primary-700" />
            <span class="text-sm font-medium text-ink">{{
              t("home.tags.realtimeBilling")
            }}</span>
          </div>
        </div>

        <!-- Features Grid -->
        <div class="mb-12 grid gap-6 md:grid-cols-3">
          <!-- Feature 1: Unified Gateway -->
          <div
            class="group rounded-lg border border-hairline bg-surface-card p-7 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-card-hover"
          >
            <div
              class="mb-5 flex h-12 w-12 items-center justify-center rounded-lg bg-surface-dark text-on-dark transition-transform group-hover:scale-105"
            >
              <Icon name="server" size="lg" class="text-on-primary" />
            </div>
            <h2 class="mb-2 font-display text-2xl font-medium text-ink">
              {{ t("home.features.unifiedGateway") }}
            </h2>
            <p class="text-sm leading-relaxed text-body">
              {{ t("home.features.unifiedGatewayDesc") }}
            </p>
          </div>

          <!-- Feature 2: Account Pool -->
          <div
            class="group rounded-lg border border-hairline bg-surface-card p-7 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-card-hover"
          >
            <div
              class="mb-5 flex h-12 w-12 items-center justify-center rounded-lg bg-primary-500 text-on-primary transition-transform group-hover:scale-105"
            >
              <svg
                class="h-6 w-6 text-on-primary"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z"
                />
              </svg>
            </div>
            <h2 class="mb-2 font-display text-2xl font-medium text-ink">
              {{ t("home.features.multiAccount") }}
            </h2>
            <p class="text-sm leading-relaxed text-body">
              {{ t("home.features.multiAccountDesc") }}
            </p>
          </div>

          <!-- Feature 3: Billing & Quota -->
          <div
            class="group rounded-lg border border-hairline bg-surface-card p-7 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-card-hover"
          >
            <div
              class="mb-5 flex h-12 w-12 items-center justify-center rounded-lg bg-accent-amber text-on-primary transition-transform group-hover:scale-105"
            >
              <svg
                class="h-6 w-6 text-on-primary"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
                />
              </svg>
            </div>
            <h2 class="mb-2 font-display text-2xl font-medium text-ink">
              {{ t("home.features.balanceQuota") }}
            </h2>
            <p class="text-sm leading-relaxed text-body">
              {{ t("home.features.balanceQuotaDesc") }}
            </p>
          </div>
        </div>

        <!-- Supported Providers -->
        <div class="mb-8 text-center">
          <h2 class="mb-3 font-display text-4xl font-medium text-ink">
            {{ t("home.providers.title") }}
          </h2>
          <p class="text-sm text-muted">
            {{ t("home.providers.description") }}
          </p>
        </div>

        <div class="mb-16 flex flex-wrap items-center justify-center gap-4">
          <!-- Claude - Supported -->
          <div
            class="flex items-center gap-2 rounded-lg border border-primary-200 bg-canvas px-5 py-3 ring-1 ring-primary-500/10"
          >
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary-500"
            >
              <span class="text-xs font-bold text-on-primary">C</span>
            </div>
            <span class="text-sm font-medium text-body">{{
              t("home.providers.claude")
            }}</span>
            <span
              class="rounded bg-primary-100 px-1.5 py-0.5 text-[10px] font-medium text-primary-700"
              >{{ t("home.providers.supported") }}</span
            >
          </div>
          <!-- GPT - Supported -->
          <div
            class="flex items-center gap-2 rounded-lg border border-primary-200 bg-canvas px-5 py-3 ring-1 ring-primary-500/10"
          >
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-surface-dark"
            >
              <span class="text-xs font-bold text-on-primary">G</span>
            </div>
            <span class="text-sm font-medium text-body">GPT</span>
            <span
              class="rounded bg-primary-100 px-1.5 py-0.5 text-[10px] font-medium text-primary-700"
              >{{ t("home.providers.supported") }}</span
            >
          </div>
          <!-- Gemini - Supported -->
          <div
            class="flex items-center gap-2 rounded-lg border border-primary-200 bg-canvas px-5 py-3 ring-1 ring-primary-500/10"
          >
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-accent-amber"
            >
              <span class="text-xs font-bold text-on-primary">G</span>
            </div>
            <span class="text-sm font-medium text-body">{{
              t("home.providers.gemini")
            }}</span>
            <span
              class="rounded bg-primary-100 px-1.5 py-0.5 text-[10px] font-medium text-primary-700"
              >{{ t("home.providers.supported") }}</span
            >
          </div>
          <!-- Antigravity - Supported -->
          <div
            class="flex items-center gap-2 rounded-lg border border-primary-200 bg-canvas px-5 py-3 ring-1 ring-primary-500/10"
          >
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-error"
            >
              <span class="text-xs font-bold text-on-primary">A</span>
            </div>
            <span class="text-sm font-medium text-body">{{
              t("home.providers.antigravity")
            }}</span>
            <span
              class="rounded bg-primary-100 px-1.5 py-0.5 text-[10px] font-medium text-primary-700"
              >{{ t("home.providers.supported") }}</span
            >
          </div>
          <!-- More - Coming Soon -->
          <div
            class="flex items-center gap-2 rounded-lg border border-dashed border-hairline bg-cream px-5 py-3"
          >
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-muted"
            >
              <span class="text-xs font-bold text-on-primary">+</span>
            </div>
            <span class="text-sm font-medium text-body">{{
              t("home.providers.more")
            }}</span>
            <span
              class="rounded bg-canvas px-1.5 py-0.5 text-[10px] font-medium text-body-strong"
              >{{ t("home.providers.soon") }}</span
            >
          </div>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer
      class="relative z-10 border-t border-hairline bg-cream/55 px-6 py-8"
    >
      <div
        class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-4 text-center sm:flex-row sm:text-left"
      >
        <p class="text-sm text-muted">
          &copy; {{ currentYear }} {{ siteName }}.
          {{ t("home.footer.allRightsReserved") }}
        </p>
        <div class="flex items-center gap-4">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm text-muted transition-colors hover:text-ink"
          >
            {{ t("home.docs") }}
          </a>
          <a
            :href="githubUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm text-muted transition-colors hover:text-ink"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useAuthStore, useAppStore } from "@/stores";
import LocaleSwitcher from "@/components/common/LocaleSwitcher.vue";
import Icon from "@/components/icons/Icon.vue";

const { t } = useI18n();

const authStore = useAuthStore();
const appStore = useAppStore();

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(
  () =>
    appStore.cachedPublicSettings?.site_name || appStore.siteName || "Sub2API",
);
const siteLogo = computed(
  () => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || "",
);
const siteSubtitle = computed(
  () =>
    appStore.cachedPublicSettings?.site_subtitle || "AI API Gateway Platform",
);
const docUrl = computed(
  () => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || "",
);
const homeContent = computed(
  () => appStore.cachedPublicSettings?.home_content || "",
);

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim();
  return content.startsWith("http://") || content.startsWith("https://");
});

// GitHub URL
const githubUrl = "https://github.com/Wei-Shaw/sub2api";

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated);
const isAdmin = computed(() => authStore.isAdmin);
const dashboardPath = computed(() =>
  isAdmin.value ? "/admin/dashboard" : "/dashboard",
);
const userInitial = computed(() => {
  const user = authStore.user;
  if (!user || !user.email) return "";
  return user.email.charAt(0).toUpperCase();
});
const shellUsername = computed(() => {
  const username = authStore.user?.username?.trim();
  return username || "guest";
});

type GeoStatus = "loading" | "success" | "error";

interface GeoInfo {
  ip: string;
  asnText: string;
  countryName: string;
}

const GEO_REQUEST_TIMEOUT_MS = 2500;
const geoStatus = ref<GeoStatus>("loading");
const geoInfo = ref<GeoInfo>({
  ip: "",
  asnText: "",
  countryName: "",
});

const geoEndpoints = [
  "https://api.ip.sb/geoip",
  "https://api.country.is/?fields=asn",
] as const;

function getRecord(value: unknown): Record<string, unknown> | null {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return null;
}

function getString(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function getNumberOrString(value: unknown): string {
  if (typeof value === "number" && Number.isFinite(value)) {
    return String(value);
  }
  return getString(value);
}

function countryCodeToName(countryCode: string): string {
  if (!countryCode) {
    return "";
  }

  try {
    return (
      new Intl.DisplayNames(["en"], { type: "region" }).of(
        countryCode.toUpperCase(),
      ) || countryCode.toUpperCase()
    );
  } catch {
    return countryCode.toUpperCase();
  }
}

function normalizeCountryName(data: Record<string, unknown>): string {
  const countryName = getString(data.country_name);
  if (countryName) {
    return countryName;
  }

  const country = getString(data.country);
  if (country.length === 2) {
    return countryCodeToName(country);
  }

  return country || countryCodeToName(getString(data.country_code));
}

function normalizeAsnText(data: Record<string, unknown>): string {
  const nestedAsn = getRecord(data.asn);
  const asnNumber =
    getNumberOrString(data.asn) ||
    getNumberOrString(data.asn_number) ||
    getNumberOrString(nestedAsn?.number);
  const organization =
    getString(data.asn_organization) ||
    getString(data.organization) ||
    getString(data.isp) ||
    getString(nestedAsn?.organization);

  if (asnNumber && organization) {
    return `AS${asnNumber.replace(/^AS/i, "")} (${organization})`;
  }
  if (asnNumber) {
    return `AS${asnNumber.replace(/^AS/i, "")}`;
  }
  return organization || "Unknown";
}

function normalizeGeoInfo(payload: unknown): GeoInfo | null {
  const data = getRecord(payload);
  if (!data) {
    return null;
  }

  const ip = getString(data.ip);
  const countryName = normalizeCountryName(data);

  if (!ip || !countryName) {
    return null;
  }

  return {
    ip,
    asnText: normalizeAsnText(data),
    countryName,
  };
}

async function fetchGeoEndpoint(endpoint: string): Promise<GeoInfo | null> {
  const controller = new AbortController();
  const timeoutId = window.setTimeout(
    () => controller.abort(),
    GEO_REQUEST_TIMEOUT_MS,
  );

  try {
    const response = await fetch(endpoint, {
      cache: "no-store",
      signal: controller.signal,
    });
    if (!response.ok) {
      return null;
    }

    return normalizeGeoInfo(await response.json());
  } catch {
    return null;
  } finally {
    window.clearTimeout(timeoutId);
  }
}

async function loadGeoInfo(): Promise<void> {
  geoStatus.value = "loading";

  for (const endpoint of geoEndpoints) {
    const result = await fetchGeoEndpoint(endpoint);
    if (result) {
      geoInfo.value = result;
      geoStatus.value = "success";
      return;
    }
  }

  geoStatus.value = "error";
}

// Current year for footer
const currentYear = computed(() => new Date().getFullYear());

onMounted(() => {
  // Check auth state
  authStore.checkAuth();

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings();
  }

  loadGeoInfo();
});
</script>

<style scoped>
/* Terminal Container */
.terminal-container {
  position: relative;
  display: inline-block;
  max-width: 100%;
}

/* Terminal Window */
.terminal-window {
  width: 420px;
  max-width: min(420px, calc(100vw - 3rem));
  background: #181715;
  border-radius: 12px;
  box-shadow:
    0 25px 50px -12px rgba(20, 20, 19, 0.4),
    0 0 0 1px rgba(250, 249, 245, 0.1),
    inset 0 1px 0 rgba(250, 249, 245, 0.1);
  overflow: hidden;
  transform: perspective(1000px) rotateX(1.5deg) rotateY(-1.5deg);
  transition: transform 0.3s ease;
}

.terminal-window:hover {
  transform: perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(-4px);
}

/* Terminal Header */
.terminal-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: rgba(37, 35, 32, 0.9);
  border-bottom: 1px solid rgba(250, 249, 245, 0.05);
}

.terminal-buttons {
  display: flex;
  gap: 8px;
}

.terminal-buttons span {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.btn-close {
  background: #c64545;
}
.btn-minimize {
  background: #d4a017;
}
.btn-maximize {
  background: #5db872;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 12px;
  font-family: ui-monospace, monospace;
  color: #a09d96;
  margin-right: 52px;
}

/* Terminal Body */
.terminal-body {
  padding: 20px 24px;
  font-family: ui-monospace, "Fira Code", monospace;
  font-size: 14px;
  line-height: 2;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  opacity: 0;
  animation: line-appear 0.5s ease forwards;
}
.motd-row {
  display: grid;
  grid-template-columns: 76px minmax(0, 1fr);
  align-items: start;
  column-gap: 8px;
}

.line-1 {
  animation-delay: 0.3s;
}
.line-2 {
  animation-delay: 1s;
}
.line-3 {
  animation-delay: 1.8s;
}
.line-4 {
  animation-delay: 2.5s;
}
.line-5 {
  animation-delay: 3.1s;
}
.line-6 {
  animation-delay: 3.7s;
}
.line-cursor {
  animation-delay: 4.3s;
}

@keyframes line-appear {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.code-prompt {
  color: #5db872;
  font-weight: bold;
}
.code-path {
  color: #a09d96;
}
.code-cmd {
  color: #e8a55a;
}
.code-comment {
  color: #a09d96;
  font-style: italic;
}
.code-welcome {
  color: #faf9f5;
}
.code-success {
  color: #5db872;
  background: rgba(93, 184, 114, 0.15);
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
}
.code-label {
  color: #cc785c;
  font-weight: 600;
  white-space: nowrap;
}
.code-value {
  color: #e8a55a;
  min-width: 0;
  white-space: normal;
}
.code-value-wrap {
  overflow-wrap: break-word;
  word-break: normal;
}
.code-error {
  color: #c64545;
  font-weight: 600;
}

/* Blinking Cursor */
.cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: #cc785c;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0;
  }
}
</style>
