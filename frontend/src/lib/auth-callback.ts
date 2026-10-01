/**
 * Post-login destination helpers.
 *
 * Only relative same-origin paths are accepted so an open redirect cannot be
 * smuggled in via ?callbackURL=.
 */

export function safeCallbackURL(
  raw: string | null | undefined,
  fallback = "/",
): string {
  if (!raw) return fallback;
  // Relative path only — reject protocol-relative and absolute URLs.
  if (!raw.startsWith("/") || raw.startsWith("//")) return fallback;
  if (raw.includes("\\") || raw.includes("://")) return fallback;
  return raw;
}

/** Build /login?callbackURL=… when returnTo is a safe deep link. */
export function loginPathWithCallback(returnTo: string): string {
  const safe = safeCallbackURL(returnTo);
  if (safe === "/") return "/login";
  return `/login?callbackURL=${encodeURIComponent(safe)}`;
}
