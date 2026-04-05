import { cookies } from "next/headers";

/**
 * Extracts the raw session token from the Better Auth cookie.
 * Better Auth signs cookies as "token.hmac_signature" — we strip the
 * signature before sending to the Go backend (which does its own DB lookup).
 */
export async function getSessionToken(): Promise<string | null> {
  const cookieStore = await cookies();
  const rawCookie =
    cookieStore.get("better-auth.session_token")?.value ??
    cookieStore.get("__Secure-better-auth.session_token")?.value;

  if (!rawCookie) return null;

  // Strip HMAC signature: "token.signature" → "token"
  // Token itself may contain dots (UUID format doesn't, but future-proof)
  const parts = rawCookie.split(".");
  return parts.length > 1 ? parts.slice(0, -1).join(".") : rawCookie;
}
