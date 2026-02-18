import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, act } from "@testing-library/react";
import { UsernameForm } from "../username-form";

vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return {
    ...actual,
    apiGet: vi.fn(),
    apiPost: vi.fn(),
  };
});

import { apiGet, apiPost, ApiError } from "@/lib/api";

const mockApiGet = apiGet as ReturnType<typeof vi.fn>;
const mockApiPost = apiPost as ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockApiGet.mockReset();
  mockApiPost.mockReset();
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

function setup() {
  const onSuccess = vi.fn();
  render(<UsernameForm onSuccess={onSuccess} />);
  const input = screen.getByLabelText("Username");
  return { onSuccess, input };
}

function changeInput(input: HTMLElement, value: string) {
  fireEvent.change(input, { target: { value } });
}

async function typeAndDebounce(input: HTMLElement, value: string) {
  changeInput(input, value);
  await act(async () => {
    await vi.advanceTimersByTimeAsync(400);
  });
}

describe("UsernameForm", () => {
  it("does not call API when input is less than 3 chars", async () => {
    const { input } = setup();

    changeInput(input, "ab");
    await act(async () => {
      await vi.advanceTimersByTimeAsync(500);
    });

    expect(mockApiGet).not.toHaveBeenCalled();
  });

  it("debounces 400ms before checking availability", async () => {
    mockApiGet.mockResolvedValue({ available: true });
    const { input } = setup();

    changeInput(input, "test");

    // Not called yet — debounce hasn't fired
    expect(mockApiGet).not.toHaveBeenCalled();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(400);
    });

    expect(mockApiGet).toHaveBeenCalledWith("accounts/check-username", {
      username: "test",
    });
  });

  it("shows available state after successful check", async () => {
    mockApiGet.mockResolvedValue({ available: true });
    const { input } = setup();

    await typeAndDebounce(input, "goodname");

    expect(screen.getByRole("button", { name: /create inbox/i })).toBeEnabled();
  });

  it("shows unavailable state when username is taken", async () => {
    mockApiGet.mockResolvedValue({ available: false, reason: "Username reserved" });
    const { input } = setup();

    await typeAndDebounce(input, "taken");

    expect(screen.getByText("Username reserved")).toBeInTheDocument();
  });

  it("submits and calls onSuccess", async () => {
    mockApiGet.mockResolvedValue({ available: true });
    mockApiPost.mockResolvedValue({ ID: "1", Username: "newuser" });
    const { input, onSuccess } = setup();

    await typeAndDebounce(input, "newuser");

    await act(async () => {
      fireEvent.submit(screen.getByRole("button", { name: /create inbox/i }).closest("form")!);
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });

    expect(mockApiPost).toHaveBeenCalledWith("accounts", { Username: "newuser" });
    expect(onSuccess).toHaveBeenCalled();
  });

  it("handles 409 conflict by marking username as taken", async () => {
    mockApiGet.mockResolvedValue({ available: true });
    mockApiPost.mockRejectedValue(new ApiError(409, "Conflict"));
    const { input, onSuccess } = setup();

    await typeAndDebounce(input, "conflict");

    await act(async () => {
      fireEvent.submit(screen.getByRole("button", { name: /create inbox/i }).closest("form")!);
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });

    expect(screen.getByText("This username is already taken")).toBeInTheDocument();
    expect(onSuccess).not.toHaveBeenCalled();
  });

  it("displays error message for non-409 errors", async () => {
    mockApiGet.mockResolvedValue({ available: true });
    mockApiPost.mockRejectedValue(new Error("Server down"));
    const { input, onSuccess } = setup();

    await typeAndDebounce(input, "erroruser");

    await act(async () => {
      fireEvent.submit(screen.getByRole("button", { name: /create inbox/i }).closest("form")!);
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });

    expect(screen.getByText("Server down")).toBeInTheDocument();
    expect(onSuccess).not.toHaveBeenCalled();
  });
});
