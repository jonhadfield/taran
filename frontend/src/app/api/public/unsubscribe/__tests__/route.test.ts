import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { NextRequest } from "next/server";
import { GET, POST } from "../route";

const URL_BASE = "http://localhost:3002/api/public/unsubscribe";

// NextRequest has its own RequestInit whose signal is not nullable, so the
// method is all this helper needs to pass through.
function req(query: string, init?: { method?: string }) {
  return new NextRequest(`${URL_BASE}${query}`, init);
}

describe("public unsubscribe route", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("GET shows a confirmation and does not unsubscribe anyone", async () => {
    const res = await GET(req("?uid=u1&token=t1"));
    const body = await res.text();

    expect(res.status).toBe(200);
    expect(body).toContain("Stop receiving digests?");
    expect(body).toContain('method="post"');
    // Mail scanners and prefetchers follow links in email. If GET mutated,
    // people would be unsubscribed without ever clicking.
    expect(fetch).not.toHaveBeenCalled();
  });

  it("GET puts the parameters in the form action, not in the page text", async () => {
    const res = await GET(req("?uid=u1&token=%22%3E%3Cscript%3E"));
    const body = await res.text();

    expect(body).not.toContain("<script>");
    expect(body).toContain("token=%22%3E%3Cscript%3E");
  });

  it("POST forwards to the backend and passes its response through", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response("<html>You have been unsubscribed</html>", {
        status: 200,
        headers: { "content-type": "text/html; charset=utf-8" },
      }),
    );

    const res = await POST(req("?uid=u1&token=t1", { method: "POST" }));
    const body = await res.text();

    expect(res.status).toBe(200);
    expect(body).toContain("unsubscribed");

    const called = vi.mocked(fetch).mock.calls[0][0] as string;
    expect(called).toContain("/api/public/unsubscribe");
    expect(called).toContain("uid=u1");
    expect(vi.mocked(fetch).mock.calls[0][1]).toMatchObject({ method: "POST" });
  });

  it("POST keeps the backend's rejection status", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response("<html>Invalid or expired link</html>", { status: 403 }),
    );

    const res = await POST(req("?uid=u1&token=bad", { method: "POST" }));
    expect(res.status).toBe(403);
  });

  it("returns 400 when the link is incomplete", async () => {
    expect((await GET(req("?uid=u1"))).status).toBe(400);
    expect((await POST(req("", { method: "POST" }))).status).toBe(400);
    expect(fetch).not.toHaveBeenCalled();
  });

  it("returns 502 rather than an error page when the backend is unreachable", async () => {
    vi.mocked(fetch).mockRejectedValue(new Error("connection refused"));

    const res = await POST(req("?uid=u1&token=t1", { method: "POST" }));
    expect(res.status).toBe(502);
    expect(await res.text()).toContain("Something went wrong");
  });
});
