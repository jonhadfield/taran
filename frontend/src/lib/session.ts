import { createHmac } from "crypto";
import { cookies } from "next/headers";

/** Cookie names Better Auth uses for the session token, insecure and secure. */
export const SESSION_COOKIE = "better-auth.session_token";
export const SECURE_SESSION_COOKIE = "__Secure-better-auth.session_token";

/**
 * Extracts the raw session token from the Better Auth cookie.
 * Better Auth signs cookies as "token.hmac_signature" — we strip the
 * signature before sending to the Go backend (which does its own DB lookup).
 */
export async function getSessionToken(): Promise<string | null> {
  const cookieStore = await cookies();
  const rawCookie =
    cookieStore.get(SESSION_COOKIE)?.value ??
    cookieStore.get(SECURE_SESSION_COOKIE)?.value;

  if (!rawCookie) return null;

  // Strip HMAC signature: "token.signature" → "token"
  // Token itself may contain dots (UUID format doesn't, but future-proof)
  const parts = rawCookie.split(".");
  return parts.length > 1 ? parts.slice(0, -1).join(".") : rawCookie;
}

/**
 * Signs a session token the way Better Auth signs its session cookie, producing
 * the `"token.signature"` value it expects to read back.
 *
 * The signature must be the **standard base64** encoding of the raw HMAC-SHA256
 * digest. better-call's `getSignedCookie` rejects anything else before it even
 * checks the bytes: it requires a 44-character signature ending in `=`, which a
 * 64-character hex digest never satisfies. A hex signature therefore carries the
 * correct HMAC but is silently treated as an invalid session.
 *
 * @see node_modules/better-call/dist/crypto.mjs — makeSignature / verifySignature
 */
export function signSessionCookieValue(token: string, secret: string): string {
  const signature = createHmac("sha256", secret).update(token).digest("base64");
  return `${token}.${signature}`;
}
