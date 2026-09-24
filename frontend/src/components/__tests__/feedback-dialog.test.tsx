import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { FeedbackDialog } from "../feedback-dialog";

vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return { ...actual, apiPost: vi.fn() };
});
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { apiPost, ApiError } from "@/lib/api";
import { toast } from "sonner";

const mockPost = apiPost as ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockPost.mockReset();
});

function open() {
  const onOpenChange = vi.fn();
  render(<FeedbackDialog open onOpenChange={onOpenChange} />);
  return { onOpenChange, box: screen.getByLabelText("Your feedback") };
}

describe("FeedbackDialog", () => {
  it("sends the trimmed message and closes", async () => {
    mockPost.mockResolvedValue({ status: "sent" });
    const { onOpenChange, box } = open();

    fireEvent.change(box, { target: { value: "  more detail on finance please  " } });
    fireEvent.click(screen.getByRole("button", { name: "Send feedback" }));

    await waitFor(() =>
      expect(mockPost).toHaveBeenCalledWith("feedback", {
        Message: "more detail on finance please",
      }),
    );
    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(toast.success).toHaveBeenCalledWith("Thanks — your feedback is on its way");
  });

  it("will not send an empty message", () => {
    open();
    expect(screen.getByRole("button", { name: "Send feedback" })).toBeDisabled();
  });

  it("passes on the server's explanation when it declines", async () => {
    mockPost.mockRejectedValue(new ApiError(503, "feedback is not available right now"));
    const { box } = open();

    fireEvent.change(box, { target: { value: "hello" } });
    fireEvent.click(screen.getByRole("button", { name: "Send feedback" }));

    await waitFor(() =>
      expect(toast.error).toHaveBeenCalledWith("feedback is not available right now"),
    );
  });

  it("keeps a server error generic", async () => {
    mockPost.mockRejectedValue(new ApiError(500, "could not send your feedback, please try again"));
    const { box } = open();

    fireEvent.change(box, { target: { value: "hello" } });
    fireEvent.click(screen.getByRole("button", { name: "Send feedback" }));

    await waitFor(() =>
      expect(toast.error).toHaveBeenCalledWith("Couldn't send your feedback, please try again"),
    );
  });
});
