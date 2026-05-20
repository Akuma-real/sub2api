export const BILLING_MODE_TOKEN = "token";
export const BILLING_MODE_PER_REQUEST = "per_request";
export const BILLING_MODE_IMAGE = "image";

export function getBillingModeLabel(
  mode: string | null | undefined,
  t: (key: string) => string,
): string {
  switch (mode) {
    case BILLING_MODE_PER_REQUEST:
      return t("admin.usage.billingModePerRequest");
    case BILLING_MODE_IMAGE:
      return t("admin.usage.billingModeImage");
    default:
      return t("admin.usage.billingModeToken");
  }
}

export function getBillingModeBadgeClass(
  mode: string | null | undefined,
): string {
  switch (mode) {
    case BILLING_MODE_PER_REQUEST:
      return "bg-surface-card text-primary-700";
    case BILLING_MODE_IMAGE:
      return "bg-surface-card text-primary-700";
    default:
      return "bg-accent-teal/15 text-accent-teal";
  }
}
