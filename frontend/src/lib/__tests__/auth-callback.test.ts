import { describe, it, expect } from "vitest";
import { safeCallbackURL, loginPathWithCallback } from "../auth-callback";

describe("safeCallbackURL", () => {
  it("returns relative paths", () => {
    expect(safeCallbackURL("/inbox/abc")).toBe("/inbox/abc");
    expect(safeCallbackURL("/inbox/abc?tab=1")).toBe("/inbox/abc?tab=1");
  });

  it("falls back for missing or unsafe values", () => {
    expect(safeCallbackURL(null)).toBe("/");
    expect(safeCallbackURL("")).toBe("/");
    expect(safeCallbackURL("https://evil.com")).toBe("/");
    expect(safeCallbackURL("//evil.com")).toBe("/");
    expect(safeCallbackURL("/\\evil")).toBe("/");
    expect(safeCallbackURL("/inbox/http://x")).toBe("/");
  });
});

describe("loginPathWithCallback", () => {
  it("omits the param for the homepage", () => {
    expect(loginPathWithCallback("/")).toBe("/login");
  });

  it("encodes deep links", () => {
    expect(loginPathWithCallback("/inbox/abc")).toBe(
      "/login?callbackURL=%2Finbox%2Fabc",
    );
  });
});
