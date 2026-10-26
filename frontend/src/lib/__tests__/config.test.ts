import { afterEach, describe, expect, it, vi } from "vitest";
import { siteUrl } from "../config";

afterEach(() => {
  vi.unstubAllEnvs();
});

describe("siteUrl", () => {
  it("uses a public https origin from APP_URL", () => {
    vi.stubEnv("NEXT_PUBLIC_APP_URL", "https://mailbrief.io/");
    expect(siteUrl()).toBe("https://mailbrief.io");
  });

  it("does not emit localhost into canonical URLs", () => {
    vi.stubEnv("NEXT_PUBLIC_APP_URL", "http://localhost:3002");
    expect(siteUrl()).toBe("https://mailbrief.io");
  });
});
