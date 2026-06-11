import { mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import HomeView from "@/views/HomeView.vue";

const { regionState, checkAuthMock, fetchPublicSettingsMock } =
  vi.hoisted(() => ({
    regionState: {
      isRestricted: false,
      geoCountryName: "United States",
    },
    checkAuthMock: vi.fn(),
    fetchPublicSettingsMock: vi.fn(),
  }));

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string>) =>
        key === "regionRestriction.detectedFrom"
          ? `Detected access location: ${params?.country || ""}`
          : key,
    }),
  };
});

vi.mock("@/stores", () => ({
  useAuthStore: () => ({
    isAuthenticated: false,
    isAdmin: false,
    user: null,
    checkAuth: checkAuthMock,
  }),
  useAppStore: () => ({
    cachedPublicSettings: null,
    siteName: "Sub2API",
    siteLogo: "",
    docUrl: "",
    publicSettingsLoaded: true,
    fetchPublicSettings: fetchPublicSettingsMock,
  }),
}));

vi.mock("@/composables/useRegionRestriction", () => ({
  useRegionRestriction: () => ({
    geoStatus: { value: "success" },
    geoInfo: {
      __v_isRef: true,
      get value() {
        return {
          ip: "203.0.113.1",
          asnText: "AS4134",
          countryName: regionState.geoCountryName,
          countryCode: regionState.isRestricted ? "CN" : "US",
        };
      },
    },
    restrictionCopy: {
      __v_isRef: true,
      get value() {
        return {
          title: "Configured restriction title",
          message: "Configured restriction message",
          detectedText: "Detected access location: {country}",
          actionText: "Configured blocked action",
        };
      },
    },
    isMainlandChinaRestricted: {
      __v_isRef: true,
      get value() {
        return regionState.isRestricted;
      },
    },
    loadRegionRestriction: vi.fn(),
  }),
}));

describe("HomeView region restriction UI", () => {
  beforeEach(() => {
    regionState.isRestricted = false;
    regionState.geoCountryName = "United States";
    checkAuthMock.mockReset();
    fetchPublicSettingsMock.mockReset();
  });

  it("renders disabled entry buttons for Mainland China visitors", () => {
    regionState.isRestricted = true;
    regionState.geoCountryName = "China";

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          LocaleSwitcher: true,
          Icon: true,
        },
      },
    });

    expect(wrapper.text()).toContain("Configured restriction title");
    expect(wrapper.text()).toContain("Configured restriction message");
    expect(wrapper.text()).toContain("Detected access location: China");
    expect(wrapper.findAll("button[disabled]").length).toBeGreaterThanOrEqual(2);
  });

  it("keeps the normal login router link outside Mainland China", () => {
    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          LocaleSwitcher: true,
          Icon: true,
        },
      },
    });

    expect(wrapper.text()).not.toContain("Configured restriction title");
    expect(wrapper.findAll("button[disabled]")).toHaveLength(0);
  });
});
