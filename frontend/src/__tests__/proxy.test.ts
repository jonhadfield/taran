// @vitest-environment node
import { describe, it, expect } from "vitest";
import { unstable_doesMiddlewareMatch } from "next/experimental/testing/server";
import { config } from "../proxy";

const runsProxy = (path: string) =>
  unstable_doesMiddlewareMatch({ config, url: `https://mailbrief.io${path}` });

describe("proxy matcher", () => {
  it.each([
    "/",
    "/inbox",
    "/settings/delivery",
    "/login",
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
