import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { AddRuleButton, suggestRule } from "../add-rule-button";

vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return { ...actual, apiPost: vi.fn() };
});

vi.mock("next/navigation", () => ({ useRouter: () => ({ push: vi.fn() }) }));
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { apiPost } from "@/lib/api";
import { toast } from "sonner";

const mockPost = apiPost as ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockPost.mockReset();
});

describe("suggestRule", () => {
  it("lists topics", () => {
    expect(suggestRule(["ISA", "Bitcoin", "Ethereum"], "Money Weekly")).toBe(
      "I want more detail on emails about ISA, Bitcoin and Ethereum",
    );
    expect(suggestRule(["Finance"], "")).toBe("I want more detail on emails about Finance");
  });

  it("falls back to the sender, then to empty", () => {
    expect(suggestRule([" "], "Money Weekly")).toBe(
      "I want more detail on emails from Money Weekly",
    );
    expect(suggestRule([], "")).toBe("");
  });
});

describe("AddRuleButton", () => {
  it("pre-fills from topics and saves the edited rule", async () => {
    mockPost.mockResolvedValue([]);
    render(<AddRuleButton topics={["Bitcoin"]} senderName="Money Weekly" />);

    fireEvent.click(screen.getByRole("button", { name: /Add analysis rule/ }));
    const textarea = screen.getByLabelText("Analysis rule") as HTMLTextAreaElement;
    expect(textarea.value).toBe("I want more detail on emails about Bitcoin");

    fireEvent.change(textarea, { target: { value: "More detail on Bitcoin price moves" } });
    fireEvent.click(screen.getByRole("button", { name: "Save rule" }));

    await waitFor(() =>
      expect(mockPost).toHaveBeenCalledWith("analysis-rules", {
        Rule: "More detail on Bitcoin price moves",
      }),
    );
    expect(toast.success).toHaveBeenCalled();
  });
});
