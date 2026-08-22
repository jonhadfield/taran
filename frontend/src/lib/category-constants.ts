export const STATUS_OPTIONS = [
  { value: "normal", label: "Normal" },
  { value: "favorite", label: "Favorite" },
  { value: "muted", label: "Muted" },
  { value: "blocked", label: "Blocked" },
];

export const CATEGORY_OPTIONS = [
  { value: "", label: "Auto" },
  { value: "newsletter", label: "Newsletter" },
  { value: "personal", label: "Personal" },
  { value: "transactional", label: "Transactional" },
  { value: "marketing", label: "Marketing" },
  { value: "notification", label: "Notification" },
  { value: "other", label: "Other" },
];

/**
 * Category options without the "Auto" entry, for contexts
 * where an empty value is not valid (e.g. auto-archive rules).
 */
export const CATEGORY_OPTIONS_NO_AUTO = CATEGORY_OPTIONS.filter(
  (o) => o.value !== "",
);

export const ALL_CATEGORIES = CATEGORY_OPTIONS
  .filter((o) => o.value !== "")
  .map((o) => o.value);

/**
 * Category identity colours, drawn from the categorical chart slots so a badge
 * and its bar in the category chart always agree.
 *
 * Slots are assigned in fixed order and never cycled — the ordering is what
 * makes adjacent pairs separable under colour-vision deficiency, so reordering
 * these silently degrades accessibility. "other" deliberately stays neutral
 * rather than taking a slot, since it is a catch-all rather than a category.
 *
 * The tint carries identity; text keeps a neutral ink token so contrast holds
 * in both modes regardless of the slot's own contrast against the surface.
 */
export const CATEGORY_COLORS: Record<string, string> = {
  newsletter: "bg-chart-1/15 text-foreground",
  personal: "bg-chart-2/15 text-foreground",
  transactional: "bg-chart-3/15 text-foreground",
  marketing: "bg-chart-4/15 text-foreground",
  notification: "bg-chart-5/15 text-foreground",
  other: "bg-muted text-muted-foreground",
};

/** Human-readable labels for each category value, derived from CATEGORY_OPTIONS. */
export const CATEGORY_LABELS: Record<string, string> = Object.fromEntries(
  CATEGORY_OPTIONS.filter((o) => o.value !== "").map((o) => [o.value, o.label]),
);

/**
 * Solid category fills for chart marks, using the same slots as CATEGORY_COLORS
 * so a bar and its badge match.
 *
 * The previous palette (blue/green/gray/orange/yellow/gray) failed validation
 * on four counts: two categories were both grey and so indistinguishable,
 * yellow and grey fell outside the lightness band, and yellow/orange sat at
 * ΔE 14.8 — below the threshold where full-colour vision can separate them.
 */
export const CATEGORY_CHART_COLORS: Record<string, string> = {
  newsletter: "bg-chart-1",
  personal: "bg-chart-2",
  transactional: "bg-chart-3",
  marketing: "bg-chart-4",
  notification: "bg-chart-5",
  other: "bg-muted-foreground/40",
};
