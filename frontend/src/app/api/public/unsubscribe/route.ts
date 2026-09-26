import { NextRequest } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

/**
 * Unsubscribe, served from the main domain.
 *
 * This used to live on api.mailbrief.io, which was CNAMEd straight at the
 * Cloud Run hostname and so served a *.a.run.app certificate that could never
 * match it. Every digest carried a link that failed TLS, in the footer and in
 * the List-Unsubscribe header both. Cloud Run domain mappings are not offered
 * in europe-west2, so rather than stand up a load balancer for one link it now
 * runs here, where the certificate is already valid and the domain is one
 * recipients recognise.
 *
 * GET renders a confirmation. It must not unsubscribe anyone on its own:
 * mail scanners and link prefetchers follow URLs in email, and a mutating GET
 * would quietly unsubscribe people who never clicked.
 *
 * POST performs it, which is also what RFC 8058 one-click sends.
 */

function page(title: string, body: string, status: number): Response {
  return new Response(
    `<!DOCTYPE html><html><head><meta charset="utf-8">` +
      `<meta name="viewport" content="width=device-width,initial-scale=1">` +
      `<title>${title} - MailBrief</title></head>` +
      `<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;` +
      `display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#fafafa;">` +
      `<div style="text-align:center;max-width:400px;padding:32px;">${body}</div></body></html>`,
    { status, headers: { "content-type": "text/html; charset=utf-8" } },
  );
}

function backendURL(uid: string, token: string): string {
  const qs = new URLSearchParams({ uid, token });
  return `${BACKEND_URL}/api/public/unsubscribe?${qs.toString()}`;
}

export async function GET(request: NextRequest) {
  const uid = request.nextUrl.searchParams.get("uid");
  const token = request.nextUrl.searchParams.get("token");

  if (!uid || !token) {
    return page("Unsubscribe", `<h1 style="font-size:20px;">Link incomplete</h1>
      <p style="color:#6b7280;">This unsubscribe link is missing information. Please use the link from a recent digest.</p>`, 400);
  }

  // The values go into the form action, URL-encoded, and are never written
  // into the HTML as text — they arrive from the query string and would
  // otherwise be an injection point.
  const action = `/api/public/unsubscribe?${new URLSearchParams({ uid, token }).toString()}`;

  return page("Unsubscribe", `<h1 style="font-size:20px;margin-bottom:8px;">Stop receiving digests?</h1>
    <p style="color:#6b7280;margin-bottom:24px;">You will no longer get MailBrief digest emails. Your account and inbox stay as they are.</p>
    <form method="post" action="${action}">
      <button type="submit" style="background:#4f46e5;color:#fff;border:0;border-radius:8px;padding:10px 20px;font-size:15px;cursor:pointer;">
        Unsubscribe
      </button>
    </form>`, 200);
}

export async function POST(request: NextRequest) {
  const uid = request.nextUrl.searchParams.get("uid");
  const token = request.nextUrl.searchParams.get("token");

  if (!uid || !token) {
    return page("Unsubscribe", `<h1 style="font-size:20px;">Link incomplete</h1>
      <p style="color:#6b7280;">This unsubscribe link is missing information.</p>`, 400);
  }

  try {
    const res = await fetch(backendURL(uid, token), { method: "POST" });
    // The backend already renders a branded result page, so pass it straight
    // through rather than second-guessing the outcome.
    return new Response(await res.text(), {
      status: res.status,
      headers: { "content-type": res.headers.get("content-type") || "text/html; charset=utf-8" },
    });
  } catch {
    return page("Unsubscribe", `<h1 style="font-size:20px;">Something went wrong</h1>
      <p style="color:#6b7280;">We could not process that just now. Please try again shortly.</p>`, 502);
  }
}
