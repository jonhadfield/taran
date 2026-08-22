import { describe, it, expect } from "vitest";
import { CATEGORY_COLORS, CATEGORY_CHART_COLORS } from "../category-constants";
import { tokenBarColor, tokenTextColor } from "../utils";

/**
 * These guard the design-token boundary. Raw palette utilities (bg-red-500)
 * bypass the token layer: they ignore the user's accent theme and are only
 * dark-mode aware if someone remembered a `dark:` variant — which, before this
 * was consolidated, held for 27 of 103 usages.
 *
 * Deliberate exceptions live elsewhere and are documented at their definition:
 * theme-picker swatches, the avatar identity hash, and user-chosen label
 * colours are all literal by design.
 */
const RAW_PALETTE =
  /\b(?:bg|text|border|ring)-(?:red|blue|green|yellow|amber|purple|violet|rose|orange|emerald|sky|indigo|pink|teal|cyan|lime|gray|slate|zinc)-\d{2,3}\b/;

describe("category colours use design tokens", () => {
  it.each(Object.entries(CATEGORY_COLORS))(
    "CATEGORY_COLORS.%s avoids raw palette utilities",
    (_category, classes) => {
      expect(classes).not.toMatch(RAW_PALETTE);
    },
  );

  it.each(Object.entries(CATEGORY_CHART_COLORS))(
    "CATEGORY_CHART_COLORS.%s avoids raw palette utilities",
    (_category, classes) => {
      expect(classes).not.toMatch(RAW_PALETTE);
    },
  );

  it("assigns each real category its own categorical slot", () => {
    const slots = Object.entries(CATEGORY_CHART_COLORS)
      .filter(([category]) => category !== "other")
      .map(([, classes]) => classes);
    // No two categories may share a slot — that was the original defect, where
    // transactional and other were both grey and so indistinguishable.
    expect(new Set(slots).size).toBe(slots.length);
  });

  it("keeps badge and chart colours on the same slot per category", () => {
    for (const category of Object.keys(CATEGORY_CHART_COLORS)) {
      const chart = CATEGORY_CHART_COLORS[category];
      const badge = CATEGORY_COLORS[category];
      const slot = chart?.match(/chart-(\d)/)?.[1];
      if (slot) expect(badge).toContain(`chart-${slot}`);
    }
  });
});

describe("usage status colours use design tokens", () => {
  it.each([0, 50, 69, 70, 89, 90, 100])("tokenBarColor(%i) is tokenised", (pct) => {
    expect(tokenBarColor(pct)).not.toMatch(RAW_PALETTE);
  });

  it.each([0, 70, 90, 100])("tokenTextColor(%i) is tokenised", (pct) => {
    expect(tokenTextColor(pct)).not.toMatch(RAW_PALETTE);
  });

  it("escalates from neutral through warning to destructive", () => {
    expect(tokenBarColor(10)).toBe("bg-primary");
    expect(tokenBarColor(75)).toBe("bg-warning");
    expect(tokenBarColor(95)).toBe("bg-destructive");
  });
});
