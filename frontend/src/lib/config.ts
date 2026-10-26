export const APP_NAME = process.env.NEXT_PUBLIC_APP_NAME || "MailBrief";

/** When false, public pages show auth/reading only — no pitch, process, or demo digest. */
export const SHOW_MARKETING =
  process.env.NEXT_PUBLIC_SHOW_MARKETING !== "false";

export const SITE_TITLE = `${APP_NAME} — AI newsletter digests`;

export const SITE_DESCRIPTION =
  "Forward your newsletters to a managed inbox and get one AI-written daily brief instead of a crowded inbox.";

/** Canonical public origin. Localhost APP_URL must not leak into robots/sitemap/OG. */
export function siteUrl(): string {
  const raw = process.env.NEXT_PUBLIC_APP_URL?.trim().replace(/\/$/, "");
  if (
    raw &&
    /^https:\/\//i.test(raw) &&
    !/localhost|127\.0\.0\.1/i.test(raw)
  ) {
    return raw;
  }
  return "https://mailbrief.io";
}

