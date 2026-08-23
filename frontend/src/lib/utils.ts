import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Short date: "24 March" */
export function formatShortDate(date: string | Date): string {
  const d = typeof date === "string" ? new Date(date) : date;
  return d.toLocaleDateString("en-GB", { day: "numeric", month: "long" });
}

const WORDS_PER_MINUTE = 200;

/** Estimate reading time at 200 WPM */
export function estimateReadingTime(text: string): string {
  const words = text.trim().split(/\s+/).filter(Boolean).length;
  const minutes = Math.ceil(words / WORDS_PER_MINUTE);
  if (minutes < 1) return "< 1 min read";
  return `${minutes} min read`;
}

/** Returns true if the URL uses http, https, or mailto protocol */
export function isSafeURL(url: string): boolean {
  try {
    const parsed = new URL(url, "https://placeholder.invalid");
    return parsed.protocol === "https:" || parsed.protocol === "http:" || parsed.protocol === "mailto:";
  } catch {
    return false;
  }
}

/** Format a token count with K/M suffixes */
export function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return n.toString();
}

/** Tailwind background class for a token-usage progress bar, by percent used */
export function tokenBarColor(percent: number): string {
  if (percent >= 90) return "bg-destructive";
  if (percent >= 70) return "bg-warning";
  return "bg-primary";
}

/** Tailwind text class for a token-usage figure, by percent used */
export function tokenTextColor(percent: number): string {
  if (percent >= 90) return "text-destructive";
  if (percent >= 70) return "text-warning";
  return "";
}

/** Pluralize a word based on count */
export function pluralize(count: number, singular: string, plural?: string) {
  return count === 1 ? singular : (plural ?? singular + "s");
}

/** Full date+time: "2026/03/24 14:30" */
export function formatDateTime(date: string | Date): string {
  const d = typeof date === "string" ? new Date(date) : date;
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  const h = String(d.getHours()).padStart(2, "0");
  const min = String(d.getMinutes()).padStart(2, "0");
  return `${y}/${m}/${day} ${h}:${min}`;
}
