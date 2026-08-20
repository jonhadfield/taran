import { describe, it, expect, vi } from "vitest";
import { createHmac, webcrypto } from "crypto";
import { signSessionCookieValue } from "../session";

vi.mock("next/headers", () => ({ cookies: vi.fn() }));

const SECRET = "test-better-auth-secret";
const TOKEN = "9f8c1d2e-4b6a-4c3d-8e7f-1a2b3c4d5e6f";

/**
 * Faithful reimplementation of better-call's `getSignedCookie`, which is what
 * Better Auth uses to read the session cookie.
 * @see node_modules/better-call/dist/context.mjs — getSignedCookie
 * @see node_modules/better-call/dist/crypto.mjs — verifySignature
 */
async function verifyAsBetterAuthWould(cookieValue: string, secret: string) {
  const signatureStartPos = cookieValue.lastIndexOf(".");
  if (signatureStartPos < 1) return false;

  const signedValue = cookieValue.substring(0, signatureStartPos);
  const signature = cookieValue.substring(signatureStartPos + 1);

  // better-call rejects the cookie outright unless the signature is a
  // 44-character base64 digest ending in "=" — before verifying any bytes.
  if (signature.length !== 44 || !signature.endsWith("=")) return false;

  const key = await webcrypto.subtle.importKey(
    "raw",
    new TextEncoder().encode(secret),
    { name: "HMAC", hash: "SHA-256" },
    false,
    ["sign", "verify"],
  );
  const binary = atob(signature);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);

  return webcrypto.subtle.verify(
    { name: "HMAC", hash: "SHA-256" },
    key,
    bytes,
    new TextEncoder().encode(signedValue),
  );
}

describe("signSessionCookieValue", () => {
  it("produces a cookie Better Auth accepts", async () => {
    const cookieValue = signSessionCookieValue(TOKEN, SECRET);
    await expect(verifyAsBetterAuthWould(cookieValue, SECRET)).resolves.toBe(true);
  });

  it("emits a 44-character base64 signature, not hex", () => {
    const signature = signSessionCookieValue(TOKEN, SECRET).split(".").pop()!;
    expect(signature).toHaveLength(44);
    expect(signature.endsWith("=")).toBe(true);
  });

  it("keeps the token recoverable by stripping the last dot-segment", () => {
    const cookieValue = signSessionCookieValue(TOKEN, SECRET);
    expect(cookieValue.split(".").slice(0, -1).join(".")).toBe(TOKEN);
  });

  it("rejects a signature signed with the wrong secret", async () => {
    const cookieValue = signSessionCookieValue(TOKEN, "a-different-secret");
    await expect(verifyAsBetterAuthWould(cookieValue, SECRET)).resolves.toBe(false);
  });

  // Regression: the proxy route previously used .digest("hex"), which carries
  // the same HMAC bytes but fails better-call's length/padding guard, silently
  // logging users out on the first token rotation.
  it("a hex signature would be rejected by Better Auth", async () => {
    const hexSignature = createHmac("sha256", SECRET).update(TOKEN).digest("hex");
    await expect(
      verifyAsBetterAuthWould(`${TOKEN}.${hexSignature}`, SECRET),
    ).resolves.toBe(false);
  });
});
