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

export const CATEGORY_COLORS: Record<string, string> = {
  newsletter: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400",
  personal: "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400",
  transactional: "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400",
  marketing: "bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400",
  notification: "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400",
  other: "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400",
};

/**
 * Simpler category colors for charts/bars (solid backgrounds only).
 */
export const CATEGORY_CHART_COLORS: Record<string, string> = {
  newsletter: "bg-blue-500",
  personal: "bg-green-500",
  transactional: "bg-gray-400",
  marketing: "bg-orange-500",
  notification: "bg-yellow-500",
  other: "bg-gray-300 dark:bg-gray-600",
};
