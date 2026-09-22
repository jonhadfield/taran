import { serverFetch } from "@/lib/server-api";
import type { AccessCheck } from "@/types/api";

export type AccessState =
  | "allowed"
  | "denied" // the backend said this user has no access
  | "unauthenticated" // session missing, expired or rotated
  | "unavailable"; // the check itself failed, so access is unknown

/** Transient failures worth one retry: rate limiting and server-side errors. */
function isTransient(message: string): boolean {
  return /API error: (429|5\d\d)/.test(message) || !/API error: \d+/.test(message);
}

/**
 * Asks the backend whether the signed-in user may use the app.
 *
 * A failed check is reported as "unavailable" rather than "denied": telling a
 * legitimate user they aren't invited because the backend was briefly
 * rate-limited or erroring is worse than showing a try-again page.
 */
export async function checkAccess(retryDelayMs = 250): Promise<AccessState> {
  for (let attempt = 0; ; attempt++) {
    try {
      const access = await serverFetch<AccessCheck>("access");
      return access.hasAccess ? "allowed" : "denied";
    } catch (err) {
      const message = err instanceof Error ? err.message : "";
      if (message.includes("Not authenticated") || message.includes("API error: 401")) {
        return "unauthenticated";
      }
      if (attempt === 0 && isTransient(message)) {
        await new Promise((r) => setTimeout(r, retryDelayMs));
        continue;
      }
      return "unavailable";
    }
  }
}
