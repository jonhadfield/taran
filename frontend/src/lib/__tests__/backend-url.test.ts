import { describe, it, expect } from "vitest";
import { buildBackendURL } from "../backend-url";

const BACKEND = "http://localhost:8080";

describe("buildBackendURL", () => {
  it("builds a normal proxied path", () => {
    expect(buildBackendURL(["emails"], BACKEND)?.toString()).toBe(
      "http://localhost:8080/api/emails",
    );
  });

  it("preserves nested resource paths", () => {
    expect(
      buildBackendURL(["emails", "abc-123", "labels", "def-456"], BACKEND)?.pathname,
    ).toBe("/api/emails/abc-123/labels/def-456");
  });

  // Regression: `new URL()` resolves dot segments, so these previously escaped
  // the /api/ prefix and reached backend routes the proxy must not expose.
  it.each([
    [["..", "..", "webhook", "email"]],
    [["%2e%2e", "%2e%2e", "webhook", "email"]],
    [[".", "..", "health"]],
    [["emails", "..", "..", "webhook", "email"]],
    [["x/../../webhook"]],
  ])("contains traversal via %j", (segments) => {
    const url = buildBackendURL(segments, BACKEND);
    // Either rejected outright, or neutralised into a path still under /api/.
    // Both are safe; what must never happen is reaching a non-/api route.
    if (url === null) return;
    expect(url.pathname.startsWith("/api/")).toBe(true);
    expect(url.origin).toBe(BACKEND);
  });

  it("rejects the dot-segment escape that reached /webhook/email", () => {
    expect(buildBackendURL(["..", "..", "webhook", "email"], BACKEND)).toBeNull();
    expect(buildBackendURL(["..", "health"], BACKEND)).toBeNull();
  });

  it("stops a segment from grafting on a query string", () => {
    const url = buildBackendURL(["emails?admin=true"], BACKEND);
    expect(url?.search).toBe("");
  });

  it("stops a segment from redirecting to another host", () => {
    expect(buildBackendURL(["..", "..", "evil.example.com"], BACKEND)).toBeNull();
    expect(buildBackendURL(["emails", "..", "..", "evil.example.com"], BACKEND)).toBeNull();
  });
});
