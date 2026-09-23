import { headers } from "next/headers";

/**
 * Header the backend reads to learn the visitor's address. The backend trusts
 * it only from callers holding the API key, i.e. this app.
 */
export const CLIENT_IP_HEADER = "X-Client-IP";

/**
 * The visitor's IP, as seen by the hosting platform.
 *
 * Every browser call reaches the backend through this app, so without it the
 * backend would see one of the platform's egress addresses for every user and
 * rate-limit them all against a single bucket.
 */
export function clientIPFromHeaders(incoming: Headers): string | null {
  // x-forwarded-for is "client, proxy1, proxy2"; the first entry is the client.
  const forwarded = incoming.get("x-forwarded-for");
  if (forwarded) {
    const first = forwarded.split(",")[0].trim();
    if (first) return first;
  }
  const real = incoming.get("x-real-ip");
  return real ? real.trim() : null;
}

/**
 * Same, for server components and route handlers without a request object.
 *
 * Returns nothing when the request headers aren't available, so a call can
 * never fail for want of an address; the backend then falls back to the
 * connecting address, as it did before.
 */
export async function clientIPHeader(): Promise<Record<string, string>> {
  try {
    const ip = clientIPFromHeaders(await headers());
    return ip ? { [CLIENT_IP_HEADER]: ip } : {};
  } catch {
    return {};
  }
}
