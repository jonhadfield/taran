import { cookies } from "next/headers";

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";
const API_KEY = process.env.API_KEY || "";

export async function serverFetch<T>(path: string): Promise<T> {
  const cookieStore = await cookies();
  const rawCookie =
    cookieStore.get("better-auth.session_token")?.value ??
    cookieStore.get("__Secure-better-auth.session_token")?.value;

  if (!rawCookie) {
    throw new Error("Not authenticated");
  }

  // Better Auth signs cookies as "token.hmac_signature" — strip the signature
  const parts = rawCookie.split(".");
  const sessionToken = parts.length > 1 ? parts.slice(0, -1).join(".") : rawCookie;

  const res = await fetch(`${BACKEND_URL}/api/${path}`, {
    headers: {
      Authorization: `Bearer ${sessionToken}`,
      "X-API-Key": API_KEY,
    },
    cache: "no-store",
  });

  if (!res.ok) {
    throw new Error(`API error: ${res.status}`);
  }

  return res.json();
}
