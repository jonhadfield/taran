import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { usePolling } from "../use-polling";

vi.mock("@/lib/api", () => ({
  apiGet: vi.fn(),
}));

import { apiGet } from "@/lib/api";

const mockApiGet = apiGet as ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockApiGet.mockReset();
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("usePolling", () => {
  it("returns initialData before fetch resolves", () => {
    mockApiGet.mockReturnValue(new Promise(() => {})); // never resolves
    const { result } = renderHook(() => usePolling("/test", { count: 0 }));
    expect(result.current.data).toEqual({ count: 0 });
    expect(result.current.loading).toBe(true);
    expect(result.current.error).toBeNull();
  });

  it("calls apiGet immediately on mount", () => {
    mockApiGet.mockResolvedValue({ count: 5 });
    renderHook(() => usePolling("/test", { count: 0 }));
    expect(mockApiGet).toHaveBeenCalledWith("/test");
  });

  it("updates data when fetch resolves", async () => {
    mockApiGet.mockResolvedValue({ count: 5 });
    const { result } = renderHook(() => usePolling("/test", { count: 0 }, 60_000));

    await act(async () => {});

    expect(result.current.data).toEqual({ count: 5 });
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it("polls at the specified interval", async () => {
    mockApiGet.mockResolvedValue({ count: 1 });
    renderHook(() => usePolling("/test", { count: 0 }, 5000));

    await act(async () => {});
    const initialCalls = mockApiGet.mock.calls.length;

    // Advance one interval
    mockApiGet.mockResolvedValue({ count: 2 });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5000);
    });

    expect(mockApiGet).toHaveBeenCalledTimes(initialCalls + 1);
  });

  it("cleans up on unmount", async () => {
    mockApiGet.mockResolvedValue({ count: 1 });
    const { unmount } = renderHook(() => usePolling("/test", { count: 0 }, 5000));

    await act(async () => {});
    unmount();

    const callCount = mockApiGet.mock.calls.length;
    await act(async () => {
      await vi.advanceTimersByTimeAsync(10000);
    });

    expect(mockApiGet).toHaveBeenCalledTimes(callCount);
  });

  it("tracks error state on fetch failure", async () => {
    mockApiGet.mockResolvedValueOnce({ count: 5 });
    const { result } = renderHook(() => usePolling("/test", { count: 0 }, 5000));

    await act(async () => {});
    expect(result.current.data).toEqual({ count: 5 });
    expect(result.current.error).toBeNull();

    mockApiGet.mockRejectedValueOnce(new Error("network error"));
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5000);
    });

    // Keeps stale data but tracks error
    expect(result.current.data).toEqual({ count: 5 });
    expect(result.current.error).toBe("network error");
  });

  it("clears error on successful fetch", async () => {
    mockApiGet.mockRejectedValueOnce(new Error("fail"));
    const { result } = renderHook(() => usePolling("/test", { count: 0 }, 5000));

    await act(async () => {});
    expect(result.current.error).toBe("fail");

    mockApiGet.mockResolvedValueOnce({ count: 10 });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5000);
    });

    expect(result.current.data).toEqual({ count: 10 });
    expect(result.current.error).toBeNull();
  });
});
