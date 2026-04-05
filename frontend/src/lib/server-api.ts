import { getSessionToken } from "@/lib/session";

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";
const API_KEY = process.env.API_KEY!; // Validated at startup — required env var

export async function serverFetch<T>(path: string): Promise<T> {
  const sessionToken = await getSessionToken();
  if (!sessionToken) {
    throw new Error("Not authenticated");
  }

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
