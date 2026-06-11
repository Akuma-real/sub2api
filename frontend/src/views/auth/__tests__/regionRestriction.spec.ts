import { flushPromises, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import LoginView from "@/views/auth/LoginView.vue";
import RegisterView from "@/views/auth/RegisterView.vue";

const {
  regionState,
  showWarningMock,
  fetchPublicSettingsMock,
} = vi.hoisted(() => ({
  regionState: {
    isRestricted: false,
    geoCountryName: "",
  },
  showWarningMock: vi.fn(),
  fetchPublicSettingsMock: vi.fn(),
}));

vi.mock("vue-router", () => ({
  useRoute: () => ({ query: {} }),
  useRouter: () => ({
    push: vi.fn(),
    currentRoute: { value: { query: {} } },
  }),
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
      locale: { value: "en" },
    }),
  };
});

vi.mock("@/stores", () => ({
  useAuthStore: () => ({
    login: vi.fn(),
    register: vi.fn(),
  }),
  useAppStore: () => ({
    fetchPublicSettings: (...args: unknown[]) => fetchPublicSettingsMock(...args),
    showWarning: (...args: unknown[]) => showWarningMock(...args),
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}));

vi.mock("@/api/auth", async () => {
  const actual = await vi.importActual<typeof import("@/api/auth")>("@/api/auth");
  return {
    ...actual,
    isTotp2FARequired: () => false,
    isWeChatWebOAuthEnabled: () => false,
    validatePromoCode: vi.fn(),
    validateInvitationCode: vi.fn(),
  };
});

vi.mock("@/composables/useRegionRestriction", () => ({
  useRegionRestriction: () => ({
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

const publicSettings = {
  turnstile_enabled: false,
  turnstile_site_key: "",
  linuxdo_oauth_enabled: false,
  dingtalk_oauth_enabled: false,
  backend_mode_enabled: false,
  oidc_oauth_enabled: false,
  oidc_oauth_provider_name: "OIDC",
  github_oauth_enabled: false,
  google_oauth_enabled: false,
  password_reset_enabled: false,
  login_agreement_enabled: false,
  login_agreement_mode: "modal",
  login_agreement_updated_at: "",
  login_agreement_revision: "",
  login_agreement_documents: [],
  registration_enabled: true,
  email_verify_enabled: false,
  promo_code_enabled: false,
  invitation_code_enabled: false,
  site_name: "Sub2API",
  registration_email_suffix_whitelist: [],
};

describe("auth region restriction UI", () => {
  beforeEach(() => {
    regionState.isRestricted = false;
    regionState.geoCountryName = "United States";
    showWarningMock.mockReset();
    fetchPublicSettingsMock.mockResolvedValue({ ...publicSettings });
  });

  it("disables login controls and renders the restriction notice", async () => {
    regionState.isRestricted = true;
    regionState.geoCountryName = "China";

    const wrapper = mount(LoginView, {
      global: {
        stubs: {
          AuthLayout: { template: "<main><slot /><slot name='footer' /></main>" },
          Icon: true,
          TurnstileWidget: true,
          TotpLoginModal: true,
        },
      },
    });
    await flushPromises();

    expect(wrapper.text()).toContain("Configured restriction title");
    expect(wrapper.text()).toContain("Configured restriction message");
    expect(wrapper.text()).toContain("Detected access location: China");
    expect(wrapper.find('input[type="email"]').attributes("disabled")).toBeDefined();
    expect(wrapper.find('input[type="password"]').attributes("disabled")).toBeDefined();
    expect(wrapper.find('button[type="submit"]').attributes("disabled")).toBeDefined();

    await wrapper.find("form").trigger("submit");
    expect(showWarningMock).toHaveBeenCalledWith("Configured blocked action");
  });

  it("keeps login controls enabled outside Mainland China", async () => {
    const wrapper = mount(LoginView, {
      global: {
        stubs: {
          AuthLayout: { template: "<main><slot /><slot name='footer' /></main>" },
          Icon: true,
          TurnstileWidget: true,
          TotpLoginModal: true,
        },
      },
    });
    await flushPromises();

    expect(wrapper.text()).not.toContain("Configured restriction title");
    expect(wrapper.find('button[type="submit"]').attributes("disabled")).toBeUndefined();
  });

  it("disables registration controls and renders the restriction notice", async () => {
    regionState.isRestricted = true;
    regionState.geoCountryName = "China";

    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          AuthLayout: { template: "<main><slot /><slot name='footer' /></main>" },
          Icon: true,
          TurnstileWidget: true,
        },
      },
    });
    await flushPromises();

    expect(wrapper.text()).toContain("Configured restriction title");
    expect(wrapper.find('input[type="email"]').attributes("disabled")).toBeDefined();
    expect(wrapper.find('button[type="submit"]').attributes("disabled")).toBeDefined();

    await wrapper.find("form").trigger("submit");
    expect(showWarningMock).toHaveBeenCalledWith("Configured blocked action");
  });
});
