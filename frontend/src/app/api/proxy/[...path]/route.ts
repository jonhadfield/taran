import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";
function requireEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`${name} environment variable is required`);
  }
  return value;
}

const API_KEY = requireEnv("API_KEY");

async function proxyRequest(request: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  const { path } = await params;

  const backendPath = `/api/${path.join("/")}`;
  const url = new URL(backendPath, BACKEND_URL);

  request.nextUrl.searchParams.forEach((value, key) => {
    url.searchParams.set(key, value);
  });

  const cookieStore = await cookies();
  const rawCookie =
    cookieStore.get("better-auth.session_token")?.value ??
    cookieStore.get("__Secure-better-auth.session_token")?.value;

  if (!rawCookie) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  // Better Auth signs cookies as "token.hmac_signature" — strip the signature
  const parts = rawCookie.split(".");
  const sessionToken = parts.length > 1 ? parts.slice(0, -1).join(".")  : rawCookie;

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

  return new NextResponse(data, {
    status: response.status,
    headers: responseHeaders,
  });
}

export const GET = proxyRequest;
export const POST = proxyRequest;
export const PATCH = proxyRequest;
export const PUT = proxyRequest;
export const DELETE = proxyRequest;
