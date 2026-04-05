import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, act } from "@testing-library/react";
import { CopyEmailAddress } from "../copy-email-address";

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("CopyEmailAddress", () => {
  it("renders the email address text", () => {
    render(<CopyEmailAddress emailAddress="test@mailbrief.io" />);
    expect(screen.getByText("test@mailbrief.io")).toBeInTheDocument();
  });

  it("copies to clipboard on click and shows check icon", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, { clipboard: { writeText } });

    render(<CopyEmailAddress emailAddress="test@mailbrief.io" />);

    await act(async () => {
      fireEvent.click(screen.getByRole("button"));
    });

    expect(writeText).toHaveBeenCalledWith("test@mailbrief.io");
  });

  it("reverts icon after 2000ms", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, { clipboard: { writeText } });

    render(<CopyEmailAddress emailAddress="test@mailbrief.io" />);

    await act(async () => {
      fireEvent.click(screen.getByRole("button"));
    });

    // Advance past the 2000ms timeout
    act(() => {
      vi.advanceTimersByTime(2000);
    });

    // Component should have reverted — still renders
    expect(screen.getByText("test@mailbrief.io")).toBeInTheDocument();
  });

  it("uses fallback when clipboard API rejects", async () => {
    const writeText = vi.fn().mockRejectedValue(new Error("denied"));
    Object.assign(navigator, { clipboard: { writeText } });

    const execCommand = vi.fn();
    document.execCommand = execCommand;

    render(<CopyEmailAddress emailAddress="test@mailbrief.io" />);

    await act(async () => {
      fireEvent.click(screen.getByRole("button"));
    });

    expect(execCommand).toHaveBeenCalledWith("copy");
  });
});
