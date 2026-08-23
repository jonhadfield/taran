/** Hour options for time pickers (0-23 with AM/PM labels) */
export const HOUR_OPTIONS = Array.from({ length: 24 }, (_, i) => {
  const hour = i % 12 || 12;
  const ampm = i < 12 ? "AM" : "PM";
  return { value: i, label: `${hour}:00 ${ampm}` };
});

/**
 * Label colour palette — name to Tailwind class mapping.
 *
 * These names are persisted per label, so the keys are part of the stored data
 * and must stay stable. Not semantic tokens: the user picks the colour.
 */
export const LABEL_COLORS: Record<string, string> = {
  blue: "bg-blue-500",
  green: "bg-green-500",
  red: "bg-red-500",
  purple: "bg-purple-500",
  orange: "bg-orange-500",
  yellow: "bg-yellow-500",
  pink: "bg-pink-500",
  cyan: "bg-cyan-500",
};

export function labelColorClass(color: string): string {
  return LABEL_COLORS[color] || "bg-gray-400";
}

/** Valid color themes for the app */
export const VALID_COLOR_THEMES = new Set([
  "neutral", "blue", "rose", "green", "violet", "amber",
]);

export type ColorTheme = "neutral" | "blue" | "rose" | "green" | "violet" | "amber";

/** Parse a cookie value into a validated ColorTheme */
export function parseColorTheme(value: string | undefined): ColorTheme {
  if (value && VALID_COLOR_THEMES.has(value)) return value as ColorTheme;
  return "neutral";
}
