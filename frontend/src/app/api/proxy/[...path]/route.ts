import { NextRequest, NextResponse } from "next/server";
import { buildBackendURL } from "@/lib/backend-url";
import { requireEnv } from "@/lib/env";
import {
  SECURE_SESSION_COOKIE,
  SESSION_COOKIE,
  getSessionToken,
  signSessionCookieValue,
} from "@/lib/session";

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";
const API_KEY = process.env.API_KEY!; // Validated at startup — required env var

// Required, not optional: by the time the backend sends us a rotated token it
// has already replaced the token in the database, so skipping the cookie write
// would log the user out with no way back. Fail at startup instead.
const AUTH_SECRET = requireEnv("BETTER_AUTH_SECRET");

async function proxyRequest(request: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  const { path } = await params;

  const url = buildBackendURL(path, BACKEND_URL);
  if (!url) {
    return NextResponse.json({ error: "invalid path" }, { status: 400 });
  }

  request.nextUrl.searchParams.forEach((value, key) => {
    url.searchParams.set(key, value);
  });

  const sessionToken = await getSessionToken();
  if (!sessionToken) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const headers: HeadersInit = {
    "Authorization": `Bearer ${sessionToken}`,
    "Content-Type": "application/json",
    "X-API-Key": API_KEY,
  };

  const fetchOptions: RequestInit = {
    method: request.method,
    headers,
  };

  if (request.method !== "GET" && request.method !== "HEAD") {
    const body = await request.text();
    if (body) {
      fetchOptions.body = body;
    }
  }

  let response: Response;
  try {
    response = await fetch(url.toString(), fetchOptions);
  } catch {
    return NextResponse.json(
      { error: "Backend unavailable" },
      { status: 502 }
    );
  }

  const data = await response.text();

  const responseHeaders: Record<string, string> = {
    "Content-Type": response.headers.get("Content-Type") || "application/json",
  };

  if (request.method === "GET") {
    responseHeaders["Cache-Control"] = "private, no-cache";
    const etag = response.headers.get("ETag");
    if (etag) responseHeaders["ETag"] = etag;
    const lastModified = response.headers.get("Last-Modified");
    if (lastModified) responseHeaders["Last-Modified"] = lastModified;
  }

  const nextResponse = new NextResponse(data, {
    status: response.status,
    headers: responseHeaders,
  });

  // Handle session token rotation from the backend
  const rotatedToken = response.headers.get("X-Session-Token-Rotated");
  if (rotatedToken) {
    const cookieValue = signSessionCookieValue(rotatedToken, AUTH_SECRET);
    const { name: cookieName, secure } = sessionCookieTarget(request);

    nextResponse.cookies.set(cookieName, cookieValue, {
      httpOnly: true,
      sameSite: "lax",
      secure,
      path: "/",
      maxAge: 60 * 60 * 24 * 30, // 30 days
    });
  }

  return nextResponse;
}

/**
 * Chooses which session cookie to overwrite when the backend rotates a token.
 *
 * Prefer the name the request actually carried, so we replace the cookie Better
 * Auth issued rather than creating a second one under the other name — the two
 * readers disagree about precedence, so a split would leave different parts of
 * the app looking at different tokens.
 *
 * Falling back on `request.url` alone is unreliable behind TLS termination,
 * where the proxy sees plain http internally and would drop the Secure flag;
 * x-forwarded-proto reflects what the browser actually used.
 */
function sessionCookieTarget(request: NextRequest): { name: string; secure: boolean } {
  if (request.cookies.has(SECURE_SESSION_COOKIE)) {
    return { name: SECURE_SESSION_COOKIE, secure: true };
  }
  if (request.cookies.has(SESSION_COOKIE)) {
    return { name: SESSION_COOKIE, secure: isSecureRequest(request) };
  }
  const secure = isSecureRequest(request);
  return { name: secure ? SECURE_SESSION_COOKIE : SESSION_COOKIE, secure };
}

function isSecureRequest(request: NextRequest): boolean {
  // May be a comma-separated list when proxies are chained; the first entry is
  // the protocol the client used.
  const forwardedProto = request.headers.get("x-forwarded-proto");
  if (forwardedProto) {
    return forwardedProto.split(",")[0].trim().toLowerCase() === "https";
  }
  return request.nextUrl.protocol === "https:";
}

export const GET = proxyRequest;
export const POST = proxyRequest;
export const PATCH = proxyRequest;
export const PUT = proxyRequest;
export const DELETE = proxyRequest;
