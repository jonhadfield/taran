import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/server-api", () => ({ serverFetch: vi.fn() }));

import { serverFetch } from "@/lib/server-api";
import { checkAccess } from "../access";

const mockFetch = serverFetch as ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockFetch.mockReset();
});

describe("checkAccess", () => {
  it("reports allowed and denied from the backend's answer", async () => {
    mockFetch.mockResolvedValueOnce({ hasAccess: true, reason: "invited" });
    expect(await checkAccess(0)).toBe("allowed");

    mockFetch.mockResolvedValueOnce({ hasAccess: false, reason: "not_invited" });
    expect(await checkAccess(0)).toBe("denied");
  });

  it("reports unauthenticated without retrying", async () => {
    mockFetch.mockRejectedValue(new Error("API error: 401"));
    expect(await checkAccess(0)).toBe("unauthenticated");
    expect(mockFetch).toHaveBeenCalledTimes(1);

    mockFetch.mockReset();
    mockFetch.mockRejectedValue(new Error("Not authenticated"));
    expect(await checkAccess(0)).toBe("unauthenticated");
    expect(mockFetch).toHaveBeenCalledTimes(1);
  });

  it("retries a rate-limited check and succeeds", async () => {
    mockFetch
      .mockRejectedValueOnce(new Error("API error: 429"))
      .mockResolvedValueOnce({ hasAccess: true, reason: "invited" });
    expect(await checkAccess(0)).toBe("allowed");
    expect(mockFetch).toHaveBeenCalledTimes(2);
  });

  it("reports unavailable, never denied, when the check keeps failing", async () => {
    for (const err of ["API error: 429", "API error: 500", "fetch failed"]) {
      mockFetch.mockReset();
      mockFetch.mockRejectedValue(new Error(err));
      expect(await checkAccess(0)).toBe("unavailable");
      expect(mockFetch).toHaveBeenCalledTimes(2); // one retry
    }
  });

  it("does not retry other client errors", async () => {
    mockFetch.mockRejectedValue(new Error("API error: 404"));
    expect(await checkAccess(0)).toBe("unavailable");
    expect(mockFetch).toHaveBeenCalledTimes(1);
  });
});
