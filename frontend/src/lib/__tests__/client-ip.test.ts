import { describe, it, expect } from "vitest";
import { clientIPFromHeaders } from "../client-ip";

describe("clientIPFromHeaders", () => {
  it("takes the client from the front of x-forwarded-for", () => {
    const h = new Headers({ "x-forwarded-for": "203.0.113.9, 70.41.3.18, 150.172.238.178" });
    expect(clientIPFromHeaders(h)).toBe("203.0.113.9");
  });

  it("falls back to x-real-ip", () => {
    expect(clientIPFromHeaders(new Headers({ "x-real-ip": " 203.0.113.9 " }))).toBe("203.0.113.9");
  });

  it("returns null when neither header is present or usable", () => {
    expect(clientIPFromHeaders(new Headers())).toBeNull();
    expect(clientIPFromHeaders(new Headers({ "x-forwarded-for": " " }))).toBeNull();
  });
});
