/**
 * Centralized platform color definitions.
 *
 * All components that need platform-specific styling should import from here
 * instead of defining their own color mappings.
 */

export type Platform = "anthropic" | "openai" | "antigravity" | "gemini";

// ── Badge (bg + text + border, for inline badges with border) ───────
const BADGE: Record<Platform, string> = {
  anthropic: "bg-accent-amber/10 text-warning border-accent-amber/30 ",
  openai: "bg-success/10 text-success border-success/30 ",
  antigravity: "bg-primary-500/10 text-primary-700 border-primary-500/30 ",
  gemini: "bg-accent-teal/10 text-primary-700 border-accent-teal/30 ",
};
const BADGE_DEFAULT = "bg-muted/10 text-body border-muted/30 ";

// ── Light badge (softer bg, no border) ──────────────────────────────
const BADGE_LIGHT: Record<Platform, string> = {
  anthropic: "bg-accent-amber/10 text-warning ",
  openai: "bg-success/10 text-success ",
  antigravity: "bg-primary-500/10 text-primary-700 ",
  gemini: "bg-accent-teal/10 text-primary-700 ",
};

// ── Border ──────────────────────────────────────────────────────────
const BORDER: Record<Platform, string> = {
  anthropic: "border-accent-amber/20 ",
  openai: "border-success/20 ",
  antigravity: "border-primary-500/20 ",
  gemini: "border-accent-teal/20 ",
};
const BORDER_DEFAULT = "border-hairline ";

// ── Accent bar ──────────────────────────────────────────────────────
const ACCENT_BAR: Record<Platform, string> = {
  anthropic: "bg-accent-amber",
  openai: "bg-success",
  antigravity: "bg-primary-500",
  gemini: "bg-accent-teal",
};
const ACCENT_BAR_DEFAULT = "bg-primary-500";

// ── Text (price, icon) ─────────────────────────────────────────────
const TEXT: Record<Platform, string> = {
  anthropic: "text-warning ",
  openai: "text-success ",
  antigravity: "text-primary-700 ",
  gemini: "text-primary-700 ",
};
const TEXT_DEFAULT = "text-primary-600 ";

// ── Icon (check mark etc.) ──────────────────────────────────────────
const ICON: Record<Platform, string> = {
  anthropic: "text-warning ",
  openai: "text-success ",
  antigravity: "text-primary-500 ",
  gemini: "text-accent-teal ",
};
const ICON_DEFAULT = "text-primary-500 ";

// ── Button (solid bg) ───────────────────────────────────────────────
const BUTTON: Record<Platform, string> = {
  anthropic:
    "bg-accent-amber text-on-primary hover:bg-accent-amber active:bg-accent-amber ",
  openai:
    "bg-success text-on-primary hover:bg-success active:bg-success ",
  antigravity:
    "bg-primary-500 text-on-primary hover:bg-primary-600 active:bg-primary-700 ",
  gemini:
    "bg-accent-teal text-on-primary hover:bg-primary-500 active:bg-primary-600 ",
};
const BUTTON_DEFAULT = "bg-primary-500 text-on-primary hover:bg-primary-600 ";

// ── Discount badge ──────────────────────────────────────────────────
const DISCOUNT: Record<Platform, string> = {
  anthropic: "bg-accent-amber/15 text-warning ",
  openai: "bg-success/15 text-success ",
  antigravity: "bg-primary-100 text-primary-700 ",
  gemini: "bg-accent-teal/15 text-primary-700 ",
};
const DISCOUNT_DEFAULT = "bg-error/15 text-error ";

// ── Public API ──────────────────────────────────────────────────────

function isPlatform(p: string): p is Platform {
  return (
    p === "anthropic" || p === "openai" || p === "antigravity" || p === "gemini"
  );
}

export function platformBadgeClass(p: string): string {
  return isPlatform(p) ? BADGE[p] : BADGE_DEFAULT;
}

export function platformBadgeLightClass(p: string): string {
  return isPlatform(p) ? BADGE_LIGHT[p] : BADGE_DEFAULT;
}

export function platformBorderClass(p: string): string {
  return isPlatform(p) ? BORDER[p] : BORDER_DEFAULT;
}

export function platformAccentBarClass(p: string): string {
  return isPlatform(p) ? ACCENT_BAR[p] : ACCENT_BAR_DEFAULT;
}

export function platformTextClass(p: string): string {
  return isPlatform(p) ? TEXT[p] : TEXT_DEFAULT;
}

export function platformIconClass(p: string): string {
  return isPlatform(p) ? ICON[p] : ICON_DEFAULT;
}

export function platformButtonClass(p: string): string {
  return isPlatform(p) ? BUTTON[p] : BUTTON_DEFAULT;
}

export function platformDiscountClass(p: string): string {
  return isPlatform(p) ? DISCOUNT[p] : DISCOUNT_DEFAULT;
}

export function platformLabel(p: string): string {
  switch (p) {
    case "anthropic":
      return "Anthropic";
    case "openai":
      return "OpenAI";
    case "antigravity":
      return "Antigravity";
    case "gemini":
      return "Gemini";
    default:
      return p || "API";
  }
}
