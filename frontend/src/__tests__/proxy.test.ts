// @vitest-environment node
import { describe, it, expect, vi, beforeEach } from "vitest";
import { NextRequest } from "next/server";
import { unstable_doesMiddlewareMatch } from "next/experimental/testing/server";
import { getSessionCookie } from "better-auth/cookies";
import { config, isPublicPath, proxy } from "../proxy";

vi.mock("better-auth/cookies", () => ({
  getSessionCookie: vi.fn(() => null),
}));

const sessionCookie = vi.mocked(getSessionCookie);

function request(path: string) {
  return new NextRequest(new URL(path, "https://mailbrief.io"));
}

const runsProxy = (path: string) =>
  unstable_doesMiddlewareMatch({ config, url: `https://mailbrief.io${path}` });

describe("proxy matcher", () => {
  it.each([
    "/",
    "/inbox",
    "/settings/delivery",
    "/login",
    "/home",
    "/shared/abc123",
  ])("runs for app route %s", (path) => {
    expect(runsProxy(path)).toBe(true);
  });

  // Signed-out visitors are redirected to /login, so anything the login page
  // loads must skip the proxy or it is served the login page's HTML instead.
  it.each([
    "/_next/static/chunks/main.js",
    "/_next/image",
    "/api/auth/session",
    "/favicon.ico",
    "/icon.svg",
    "/apple-icon.png",
    "/logo.svg",
    "/logo-192.png",
    "/logo-512.png",
    "/digest-flow.png",
    "/sentry-tunnel",
    "/robots.txt",
    "/sitemap.xml",
    "/.well-known/security.txt",
  ])("skips static asset or API path %s", (path) => {
    expect(runsProxy(path)).toBe(false);
  });

  it.each([
    "/favicon.icon",
    "/logo.svg/inbox",
    "/logoXsvg",
    "/icon.svgs",
  ])("only skips exact asset file names, not %s", (path) => {
    expect(runsProxy(path)).toBe(true);
  });
});

describe("isPublicPath", () => {
  it("allows signed-out access to marketing and auth routes", () => {
    expect(isPublicPath("/")).toBe(true);
    expect(isPublicPath("/home")).toBe(true);
    expect(isPublicPath("/login")).toBe(true);
    expect(isPublicPath("/shared/abc")).toBe(true);
  });

  it("keeps the dashboard behind auth", () => {
    expect(isPublicPath("/inbox")).toBe(false);
    expect(isPublicPath("/settings")).toBe(false);
  });
});

describe("proxy", () => {
  beforeEach(() => {
    sessionCookie.mockReturnValue(null);
  });

  it("rewrites signed-out / to the marketing homepage", async () => {
    const res = await proxy(request("/"));
    expect(res.headers.get("x-middleware-rewrite")).toMatch(/\/home$/);
  });

  it("does not send signed-out / to login", async () => {
    const res = await proxy(request("/"));
    expect(res.status).not.toBe(307);
    expect(res.headers.get("location")).toBeNull();
  });

  it("still redirects signed-out dashboard routes to login", async () => {
    const res = await proxy(request("/inbox/example-id"));
    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toMatch(/\/login\?callbackURL=/);
  });

  it("leaves authenticated / alone so the dashboard can render", async () => {
    sessionCookie.mockReturnValue("session");
    const res = await proxy(request("/"));
    expect(res.headers.get("x-middleware-rewrite")).toBeNull();
    expect(res.headers.get("location")).toBeNull();
  });

  it("sends signed-out /home to the canonical homepage", async () => {
    const res = await proxy(request("/home"));
    expect(res.status).toBe(308);
    expect(res.headers.get("location")).toBe("https://mailbrief.io/");
  });

  it("sends authenticated /home visitors to the dashboard", async () => {
    sessionCookie.mockReturnValue("session");
    const res = await proxy(request("/home"));
    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toBe("https://mailbrief.io/");
  });
});
