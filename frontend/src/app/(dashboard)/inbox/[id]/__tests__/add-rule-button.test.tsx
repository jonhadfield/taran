import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { AddRuleButton, suggestRule } from "../add-rule-button";

vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return { ...actual, apiPost: vi.fn() };
});

vi.mock("@/lib/reanalyse", async () => {
  const actual = await vi.importActual<typeof import("@/lib/reanalyse")>("@/lib/reanalyse");
  return { ...actual, reanalyseEmail: vi.fn() };
});

vi.mock("next/navigation", () => ({ useRouter: () => ({ push: vi.fn() }) }));
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn(), info: vi.fn(), loading: vi.fn(() => "t1") } }));

import { apiPost } from "@/lib/api";
import { reanalyseEmail } from "@/lib/reanalyse";
import { toast } from "sonner";

const mockPost = apiPost as ReturnType<typeof vi.fn>;
const mockReanalyse = reanalyseEmail as ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockPost.mockReset();
  mockReanalyse.mockReset();
});

function renderButton() {
  render(
    <AddRuleButton emailId="em-1" processedAt="t0" topics={["Bitcoin"]} senderName="Money Weekly" />,
  );
}

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
  it("pre-fills from topics, saves the edited rule and re-analyses the email", async () => {
    mockPost.mockResolvedValue([]);
    mockReanalyse.mockResolvedValue("updated");
    renderButton();

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
    await waitFor(() => expect(mockReanalyse).toHaveBeenCalledWith("em-1", "t0"));
    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith("Summary updated", { id: "t1" }),
    );
  });

  it("skips re-analysis when unticked", async () => {
    mockPost.mockResolvedValue([]);
    renderButton();

    fireEvent.click(screen.getByRole("button", { name: /Add analysis rule/ }));
    fireEvent.click(screen.getByRole("checkbox", { name: /Re-analyse this email/ }));
    fireEvent.click(screen.getByRole("button", { name: "Save rule" }));

    await waitFor(() => expect(mockPost).toHaveBeenCalled());
    await waitFor(() => expect(toast.success).toHaveBeenCalled());
    expect(mockReanalyse).not.toHaveBeenCalled();
  });

  it("does not re-analyse when saving the rule fails", async () => {
    const { ApiError } = await import("@/lib/api");
    mockPost.mockRejectedValue(new ApiError(400, "maximum 20 analysis rules allowed"));
    renderButton();

    fireEvent.click(screen.getByRole("button", { name: /Add analysis rule/ }));
    fireEvent.click(screen.getByRole("button", { name: "Save rule" }));

    await waitFor(() =>
      expect(toast.error).toHaveBeenCalledWith("maximum 20 analysis rules allowed"),
    );
    expect(mockReanalyse).not.toHaveBeenCalled();
  });
});
