import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { OpenRegistrationToggle } from "../open-registration-toggle";

vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return { ...actual, apiGet: vi.fn(), apiPatch: vi.fn() };
});
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { apiGet, apiPatch, ApiError } from "@/lib/api";
import { toast } from "sonner";

const mockGet = apiGet as ReturnType<typeof vi.fn>;
const mockPatch = apiPatch as ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockGet.mockReset();
  mockPatch.mockReset();
});

describe("OpenRegistrationToggle", () => {
  it("shows the current state and sign-up count", async () => {
    mockGet.mockResolvedValue({ openRegistration: false, signups: 1 });
    render(<OpenRegistrationToggle />);

    expect(await screen.findByText("Invite-only")).toBeInTheDocument();
    expect(screen.getByRole("switch")).not.toBeChecked();
    expect(screen.getByText(/person signed up through open registration/)).toBeInTheDocument();
  });

  it("turns open registration on", async () => {
    mockGet.mockResolvedValue({ openRegistration: false, signups: 0 });
    mockPatch.mockResolvedValue({ openRegistration: true, signups: 3 });
    render(<OpenRegistrationToggle />);

    fireEvent.click(await screen.findByRole("switch"));

    await waitFor(() =>
      expect(mockPatch).toHaveBeenCalledWith("admin/settings/open-registration", {
        OpenRegistration: true,
      }),
    );
    expect(await screen.findByText("Open to everyone")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
    expect(toast.success).toHaveBeenCalledWith("Open registration is on. Anyone can now sign up.");
  });

  it("reverts the switch when saving fails", async () => {
    mockGet.mockResolvedValue({ openRegistration: false, signups: 0 });
    mockPatch.mockRejectedValue(new ApiError(500, "failed to update open registration setting"));
    render(<OpenRegistrationToggle />);

    fireEvent.click(await screen.findByRole("switch"));

    await waitFor(() =>
      expect(toast.error).toHaveBeenCalledWith("failed to update open registration setting"),
    );
    expect(screen.getByRole("switch")).not.toBeChecked();
    expect(screen.getByText("Invite-only")).toBeInTheDocument();
  });
});
