import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import {
  resetRegionRestrictionForTest,
  useRegionRestriction,
} from "@/composables/useRegionRestriction";
import { useAppStore } from "@/stores";

function mockGeoResponse(payload: Record<string, unknown>) {
  return {
    ok: true,
    json: async () => payload,
  };
}

describe("useRegionRestriction", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    resetRegionRestrictionForTest();
    useAppStore().cachedPublicSettings = {
      region_restriction_enabled: true,
      region_restriction_countries: ["CN"],
    } as any;
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("restricts Mainland China by country code", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        mockGeoResponse({
          ip: "203.0.113.1",
          country_code: "CN",
          country_name: "China",
          asn: 4134,
          asn_organization: "China Telecom",
        }),
      ),
    );

    const region = useRegionRestriction();
    await region.loadRegionRestriction(true);

    expect(region.geoStatus.value).toBe("success");
    expect(region.geoInfo.value.countryCode).toBe("CN");
    expect(region.geoInfo.value.asnText).toBe("AS4134 (China Telecom)");
    expect(region.isMainlandChinaRestricted.value).toBe(true);
  });

  it("restricts Mainland China by country name when no country code is present", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        mockGeoResponse({
          ip: "203.0.113.2",
          country: "China",
        }),
      ),
    );

    const region = useRegionRestriction();
    await region.loadRegionRestriction(true);

    expect(region.geoStatus.value).toBe("success");
    expect(region.isMainlandChinaRestricted.value).toBe(true);
  });

  it("does not restrict non-China regions", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        mockGeoResponse({
          ip: "198.51.100.1",
          country: "US",
        }),
      ),
    );

    const region = useRegionRestriction();
    await region.loadRegionRestriction(true);

    expect(region.geoStatus.value).toBe("success");
    expect(region.geoInfo.value.countryName).toBe("United States");
    expect(region.isMainlandChinaRestricted.value).toBe(false);
  });

  it("does not restrict when the backend setting is disabled", async () => {
    useAppStore().cachedPublicSettings = {
      region_restriction_enabled: false,
      region_restriction_countries: ["CN"],
    } as any;
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        mockGeoResponse({
          ip: "203.0.113.1",
          country_code: "CN",
          country_name: "China",
        }),
      ),
    );

    const region = useRegionRestriction();
    await region.loadRegionRestriction(true);

    expect(region.geoStatus.value).toBe("success");
    expect(region.isMainlandChinaRestricted.value).toBe(false);
  });

  it("does not restrict when geo lookup fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        json: async () => ({}),
      }),
    );

    const region = useRegionRestriction();
    await region.loadRegionRestriction(true);

    expect(region.geoStatus.value).toBe("error");
    expect(region.isMainlandChinaRestricted.value).toBe(false);
  });
});
