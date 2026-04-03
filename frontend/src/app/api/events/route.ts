import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export const maxDuration = 300;

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

function requireEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`${name} environment variable is required`);
  }
  return value;
}

const API_KEY = requireEnv("API_KEY");

export async function GET() {
  const cookieStore = await cookies();
  const rawCookie =
    cookieStore.get("better-auth.session_token")?.value ??
    cookieStore.get("__Secure-better-auth.session_token")?.value;

  if (!rawCookie) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const parts = rawCookie.split(".");
  const sessionToken = parts.length > 1 ? parts.slice(0, -1).join(".") : rawCookie;

  let response: Response;
  try {
    response = await fetch(`${BACKEND_URL}/api/events`, {
      headers: {
        "Authorization": `Bearer ${sessionToken}`,
        "X-API-Key": API_KEY,
        "Accept": "text/event-stream",
      },
    });
  } catch {
    return NextResponse.json({ error: "Backend unavailable" }, { status: 502 });
  }

  if (!response.ok || !response.body) {
    return new NextResponse(response.body, { status: response.status });
  }

  return new Response(response.body, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache, no-transform",
      "Connection": "keep-alive",
      "X-Accel-Buffering": "no",
    },
  });
}
