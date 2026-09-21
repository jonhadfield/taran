import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { AnalysisRulesSettings } from "../analysis-rules-settings";
import type { AnalysisRule } from "@/types/api";

vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return {
    ...actual,
    apiGet: vi.fn(),
    apiPost: vi.fn(),
    apiPatch: vi.fn(),
    apiDelete: vi.fn(),
  };
});

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { apiGet, apiPost, apiPatch, apiDelete, ApiError } from "@/lib/api";
import { toast } from "sonner";

const mockGet = apiGet as ReturnType<typeof vi.fn>;
const mockPost = apiPost as ReturnType<typeof vi.fn>;
const mockPatch = apiPatch as ReturnType<typeof vi.fn>;
const mockDelete = apiDelete as ReturnType<typeof vi.fn>;

function rule(id: string, text: string, isActive = true): AnalysisRule {
  return { ID: id, Rule: text, IsActive: isActive, CreatedAt: "", UpdatedAt: "" };
}

beforeEach(() => {
  mockGet.mockReset();
  mockPost.mockReset();
  mockPatch.mockReset();
  mockDelete.mockReset();
});

describe("AnalysisRulesSettings", () => {
  it("shows examples when there are no rules", async () => {
    mockGet.mockResolvedValue([]);
    render(<AnalysisRulesSettings />);
    expect(await screen.findByText(/No rules yet/)).toBeInTheDocument();
    expect(screen.getByText("0/20 rules")).toBeInTheDocument();
  });

  it("adds a rule", async () => {
    mockGet.mockResolvedValue([]);
    mockPost.mockResolvedValue([rule("r1", "More detail on Bitcoin")]);
    render(<AnalysisRulesSettings />);

    fireEvent.click(await screen.findByRole("button", { name: /Add rule/ }));
    fireEvent.change(screen.getByLabelText("New rule"), {
      target: { value: "  More detail on Bitcoin  " },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() =>
      expect(mockPost).toHaveBeenCalledWith("analysis-rules", { Rule: "More detail on Bitcoin" }),
    );
    expect(await screen.findByText("More detail on Bitcoin")).toBeInTheDocument();
  });

  it("shows the server's validation message on failure", async () => {
    mockGet.mockResolvedValue([]);
    mockPost.mockRejectedValue(new ApiError(400, "maximum 20 analysis rules allowed"));
    render(<AnalysisRulesSettings />);

    fireEvent.click(await screen.findByRole("button", { name: /Add rule/ }));
    fireEvent.change(screen.getByLabelText("New rule"), { target: { value: "x" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() =>
      expect(toast.error).toHaveBeenCalledWith("maximum 20 analysis rules allowed"),
    );
  });

  it("toggles a rule off", async () => {
    mockGet.mockResolvedValue([rule("r1", "Rule one")]);
    mockPatch.mockResolvedValue([rule("r1", "Rule one", false)]);
    render(<AnalysisRulesSettings />);

    fireEvent.click(await screen.findByRole("switch", { name: "Disable rule" }));

    await waitFor(() =>
      expect(mockPatch).toHaveBeenCalledWith("analysis-rules/r1", { IsActive: false }),
    );
    expect(await screen.findByRole("switch", { name: "Enable rule" })).toBeInTheDocument();
  });

  it("edits a rule", async () => {
    mockGet.mockResolvedValue([rule("r1", "Old text")]);
    mockPatch.mockResolvedValue([rule("r1", "New text")]);
    render(<AnalysisRulesSettings />);

    fireEvent.click(await screen.findByRole("button", { name: "Edit rule" }));
    fireEvent.change(screen.getByRole("textbox", { name: "Edit rule" }), {
      target: { value: "New text" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() =>
      expect(mockPatch).toHaveBeenCalledWith("analysis-rules/r1", { Rule: "New text" }),
    );
    expect(await screen.findByText("New text")).toBeInTheDocument();
  });

  it("deletes a rule", async () => {
    mockGet.mockResolvedValue([rule("r1", "Rule one")]);
    mockDelete.mockResolvedValue(undefined);
    render(<AnalysisRulesSettings />);

    fireEvent.click(await screen.findByRole("button", { name: "Delete rule" }));

    await waitFor(() => expect(mockDelete).toHaveBeenCalledWith("analysis-rules/r1"));
    await waitFor(() => expect(screen.queryByText("Rule one")).not.toBeInTheDocument());
  });

  it("disables adding at the 20 rule limit", async () => {
    mockGet.mockResolvedValue(Array.from({ length: 20 }, (_, i) => rule(`r${i}`, `Rule ${i}`)));
    render(<AnalysisRulesSettings />);

    expect(await screen.findByRole("button", { name: /Add rule/ })).toBeDisabled();
    expect(screen.getByText("20/20 rules")).toBeInTheDocument();
  });
});
