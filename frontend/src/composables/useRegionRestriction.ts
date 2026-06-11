import { computed, ref } from "vue";
import { useAppStore } from "@/stores";

export type RegionGeoStatus = "loading" | "success" | "error";

export interface RegionGeoInfo {
  ip: string;
  asnText: string;
  countryName: string;
  countryCode: string;
}

export interface RegionRestrictionCopy {
  title: string;
  message: string;
  detectedText: string;
  actionText: string;
}

const defaultRestrictionCopy: RegionRestrictionCopy = {
  title: "当前地区暂不提供 OpenAI 相关服务",
  message:
    "根据适用法律法规、监管要求及 OpenAI 支持国家/地区政策，本站不向中国大陆地区用户提供 OpenAI/ChatGPT、API 中转、账号额度分发、共享订阅或相关付费调用服务。若你位于中国大陆，请不要注册、登录、购买或发起调用。",
  detectedText: "检测到当前访问来源：{country}",
  actionText: "当前地区暂不提供 OpenAI 相关服务，界面操作已被限制。",
};

const GEO_REQUEST_TIMEOUT_MS = 2500;

const geoStatus = ref<RegionGeoStatus>("loading");
const geoInfo = ref<RegionGeoInfo>({
  ip: "",
  asnText: "",
  countryName: "",
  countryCode: "",
});
const hasLoaded = ref(false);
let pendingLoad: Promise<void> | null = null;

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

function normalizeCountryCode(data: Record<string, unknown>): string {
  const countryCode = getString(data.country_code) || getString(data.country);
  return countryCode.length === 2 ? countryCode.toUpperCase() : "";
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

function normalizeGeoInfo(payload: unknown): RegionGeoInfo | null {
  const data = getRecord(payload);
  if (!data) {
    return null;
  }

  const ip = getString(data.ip);
  const countryCode = normalizeCountryCode(data);
  const countryName = normalizeCountryName(data);

  if (!ip || !countryName) {
    return null;
  }

  return {
    ip,
    asnText: normalizeAsnText(data),
    countryName,
    countryCode,
  };
}

function normalizeRestrictedCountries(values: unknown): string[] {
  if (!Array.isArray(values)) {
    return ["CN"];
  }

  const countries = values
    .map((value) => (typeof value === "string" ? value.trim().toUpperCase() : ""))
    .filter((value) => /^[A-Z]{2}$/.test(value));

  return countries.length > 0 ? Array.from(new Set(countries)) : ["CN"];
}

function configuredText(value: unknown, fallback: string): string {
  return typeof value === "string" && value.trim() ? value.trim() : fallback;
}

async function fetchGeoEndpoint(endpoint: string): Promise<RegionGeoInfo | null> {
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

async function loadGeoInfo(force = false): Promise<void> {
  if (pendingLoad) {
    return pendingLoad;
  }
  if (hasLoaded.value && !force) {
    return;
  }

  pendingLoad = (async () => {
    geoStatus.value = "loading";

    for (const endpoint of geoEndpoints) {
      const result = await fetchGeoEndpoint(endpoint);
      if (result) {
        geoInfo.value = result;
        geoStatus.value = "success";
        hasLoaded.value = true;
        return;
      }
    }

    geoStatus.value = "error";
    hasLoaded.value = true;
  })().finally(() => {
    pendingLoad = null;
  });

  return pendingLoad;
}

export function useRegionRestriction() {
  const appStore = useAppStore();

  const restrictionEnabled = computed(
    () => appStore.cachedPublicSettings?.region_restriction_enabled === true,
  );
  const restrictedCountries = computed(() =>
    normalizeRestrictedCountries(
      appStore.cachedPublicSettings?.region_restriction_countries,
    ),
  );
  const restrictionCopy = computed<RegionRestrictionCopy>(() => ({
    title: configuredText(
      appStore.cachedPublicSettings?.region_restriction_title,
      defaultRestrictionCopy.title,
    ),
    message: configuredText(
      appStore.cachedPublicSettings?.region_restriction_message,
      defaultRestrictionCopy.message,
    ),
    detectedText: configuredText(
      appStore.cachedPublicSettings?.region_restriction_detected,
      defaultRestrictionCopy.detectedText,
    ),
    actionText: configuredText(
      appStore.cachedPublicSettings?.region_restriction_action_text,
      defaultRestrictionCopy.actionText,
    ),
  }));

  const isRegionRestricted = computed(() => {
    if (!restrictionEnabled.value) {
      return false;
    }
    if (geoStatus.value !== "success") {
      return false;
    }
    const countryCode = geoInfo.value.countryCode.toUpperCase();
    const countryName = geoInfo.value.countryName.trim().toLowerCase();
    return (
      restrictedCountries.value.includes(countryCode) ||
      (restrictedCountries.value.includes("CN") && countryName === "china")
    );
  });

  return {
    geoStatus,
    geoInfo,
    restrictionReady: computed(() => hasLoaded.value),
    restrictionEnabled,
    restrictedCountries,
    restrictionCopy,
    isRegionRestricted,
    isMainlandChinaRestricted: isRegionRestricted,
    loadRegionRestriction: loadGeoInfo,
  };
}

export function resetRegionRestrictionForTest(): void {
  geoStatus.value = "loading";
  geoInfo.value = {
    ip: "",
    asnText: "",
    countryName: "",
    countryCode: "",
  };
  hasLoaded.value = false;
  pendingLoad = null;
}
