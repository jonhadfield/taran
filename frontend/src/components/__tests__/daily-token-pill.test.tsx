import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { DailyTokenPill } from "../daily-token-pill";
import type { UsageStats } from "@/types/api";

vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return { ...actual, apiGet: vi.fn() };
});

import { apiGet } from "@/lib/api";

const mockApiGet = apiGet as ReturnType<typeof vi.fn>;

function makeStats(over: Partial<UsageStats>): UsageStats {
  return {
    MonthlyTokensUsed: 0,
    MonthlyTokenLimit: 0,
    DailyTokensUsed: 0,
    DailyTokenLimit: 0,
    TriageTokens: 0,
    ExtractTokens: 0,
    DigestTokens: 0,
    DailyHistory: null,
    PeriodStart: "",
    PeriodEnd: "",
    ...over,
  };
}

beforeEach(() => {
  mockApiGet.mockReset();
});

describe("DailyTokenPill", () => {
  it("renders nothing when no daily limit is set", async () => {
    mockApiGet.mockResolvedValue(makeStats({ DailyTokenLimit: 0, DailyTokensUsed: 500 }));
    const { container } = render(<DailyTokenPill />);
    await waitFor(() => expect(mockApiGet).toHaveBeenCalled());
    expect(container).toBeEmptyDOMElement();
  });

  it("shows remaining tokens against the daily limit", async () => {
    mockApiGet.mockResolvedValue(makeStats({ DailyTokenLimit: 1000, DailyTokensUsed: 400 }));
    render(<DailyTokenPill />);
    expect(await screen.findByText("600 left today")).toBeInTheDocument();
  });

  it("shows a limit-reached state and never goes negative", async () => {
    mockApiGet.mockResolvedValue(makeStats({ DailyTokenLimit: 1000, DailyTokensUsed: 1500 }));
    render(<DailyTokenPill />);
    expect(await screen.findByText("Daily limit reached")).toBeInTheDocument();
  });

  it("links to settings", async () => {
    mockApiGet.mockResolvedValue(makeStats({ DailyTokenLimit: 1000, DailyTokensUsed: 100 }));
    render(<DailyTokenPill />);
    const link = await screen.findByRole("link");
    expect(link).toHaveAttribute("href", "/settings");
  });
});
