/** The prefix every proxied request must stay within on the backend. */
const API_PREFIX = "/api/";

/**
 * Builds the backend URL for a proxied `/api/proxy/<...path>` request, or
 * returns `null` if the request tries to escape the `/api/` prefix.
 *
 * `new URL()` resolves dot segments, so a path of `["..", "..", "webhook",
 * "email"]` collapses `/api/../../webhook/email` down to `/webhook/email` and
 * reaches backend routes the proxy is not meant to expose. Percent-encoded
 * forms (`%2e%2e`) normalise identically, so the guard has to test the
 * *resolved* pathname rather than the raw segments.
 *
 * Segments are also re-encoded so that a segment containing `?` or `#` cannot
 * graft a query string or fragment onto the backend URL.
 */
export function buildBackendURL(segments: string[], backendURL: string): URL | null {
  const path = segments.map(encodeURIComponent).join("/");
  const url = new URL(`${API_PREFIX}${path}`, backendURL);

  if (!url.pathname.startsWith(API_PREFIX)) return null;

  return url;
}
