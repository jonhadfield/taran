import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { requireEnv } from "@/lib/env";
import { clientIPHeader } from "@/lib/client-ip";

export const maxDuration = 300;

/** Close the proxy stream before Vercel kills it at maxDuration. */
const STREAM_TTL_MS = 270_000;

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

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

  const abort = new AbortController();
  const ttl = setTimeout(() => abort.abort(), STREAM_TTL_MS);
  let clientCancelled = false;

  let response: Response;
  try {
    response = await fetch(`${BACKEND_URL}/api/events`, {
      headers: {
        Authorization: `Bearer ${sessionToken}`,
        "X-API-Key": API_KEY,
        Accept: "text/event-stream",
        ...(await clientIPHeader()),
      },
      signal: abort.signal,
    });
  } catch {
    clearTimeout(ttl);
    if (abort.signal.aborted) {
      // TTL elapsed before the upstream connected — ask the client to retry.
      return new Response(": reconnect\n\n", {
        headers: {
          "Content-Type": "text/event-stream",
          "Cache-Control": "no-cache, no-transform",
          Connection: "keep-alive",
          "X-Accel-Buffering": "no",
        },
      });
    }
    return NextResponse.json({ error: "Backend unavailable" }, { status: 502 });
  }

  if (!response.ok || !response.body) {
    clearTimeout(ttl);
    return new NextResponse(response.body, { status: response.status });
  }

  const upstream = response.body;
  const stream = new ReadableStream<Uint8Array>({
    start(controller) {
      const reader = upstream.getReader();
      const encoder = new TextEncoder();
      let closed = false;

      const finish = (sendReconnect: boolean) => {
        if (closed) return;
        closed = true;
        clearTimeout(ttl);
        abort.signal.removeEventListener("abort", onAbort);
        try {
          if (sendReconnect && !clientCancelled) {
            controller.enqueue(encoder.encode(": reconnect\n\n"));
          }
          controller.close();
        } catch {
          // already closed
        }
        reader.cancel().catch(() => {});
      };

      const onAbort = () => finish(true);
      if (abort.signal.aborted) {
        finish(true);
        return;
      }
      abort.signal.addEventListener("abort", onAbort, { once: true });

      (async () => {
        try {
          while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            controller.enqueue(value);
          }
          finish(false);
        } catch {
          // Upstream cancelled (TTL or client disconnect) — close cleanly.
          finish(abort.signal.aborted);
        }
      })();
    },
    cancel() {
      clientCancelled = true;
      clearTimeout(ttl);
      abort.abort();
    },
  });

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache, no-transform",
      Connection: "keep-alive",
      "X-Accel-Buffering": "no",
    },
  });
}
