import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return { ...actual, apiGet: vi.fn(), apiPost: vi.fn() };
});

import { apiGet, apiPost } from "@/lib/api";
import { reanalyseEmail } from "../reanalyse";

const mockGet = apiGet as ReturnType<typeof vi.fn>;
const mockPost = apiPost as ReturnType<typeof vi.fn>;

function email(Status: string, ProcessedAt?: string) {
  return { Status, Extraction: ProcessedAt ? { ProcessedAt } : null };
}

const fast = { intervalMs: 0, attempts: 5 };

beforeEach(() => {
  mockGet.mockReset();
  mockPost.mockReset().mockResolvedValue({ status: "queued" });
});

describe("reanalyseEmail", () => {
  it("reports updated once a new extraction lands", async () => {
    mockGet
      .mockResolvedValueOnce(email("pending", "t0"))
      .mockResolvedValueOnce(email("processing", "t0"))
      .mockResolvedValueOnce(email("processed", "t1"));
    expect(await reanalyseEmail("em-1", "t0", fast)).toBe("updated");
    expect(mockPost).toHaveBeenCalledWith("emails/em-1/reprocess", {});
  });

  it("reports unchanged when processing ends with the old extraction", async () => {
    mockGet.mockResolvedValueOnce(email("processed", "t0"));
    expect(await reanalyseEmail("em-1", "t0", fast)).toBe("unchanged");
  });

  it("reports failed and timeout", async () => {
    mockGet.mockResolvedValueOnce(email("failed"));
    expect(await reanalyseEmail("em-1", undefined, fast)).toBe("failed");

    mockGet.mockReset().mockResolvedValue(email("processing", "t0"));
    expect(await reanalyseEmail("em-1", "t0", fast)).toBe("timeout");
  });
});
