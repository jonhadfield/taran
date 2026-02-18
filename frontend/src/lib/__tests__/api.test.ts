import { describe, it, expect, vi, beforeEach } from "vitest";
import { apiGet, apiPost, apiPatch, apiDelete, ApiError } from "../api";

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
});

function mockFetch(body: unknown, status = 200) {
  (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
  });
}

function mockFetchJsonError(status: number) {
  (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
    ok: false,
    status,
    json: () => Promise.reject(new Error("bad json")),
  });
}

describe("apiGet", () => {
  it("fetches the correct URL", async () => {
    mockFetch({ ok: true });
    await apiGet("emails");
    expect(fetch).toHaveBeenCalledWith("/api/proxy/emails", expect.objectContaining({
      headers: expect.objectContaining({ "Content-Type": "application/json" }),
    }));
  });

  it("appends query params", async () => {
    mockFetch({ ok: true });
    await apiGet("emails", { limit: "10", offset: "0" });
    expect(fetch).toHaveBeenCalledWith(
      "/api/proxy/emails?limit=10&offset=0",
      expect.anything(),
    );
  });
});

describe("apiPost", () => {
  it("sends POST with JSON body", async () => {
    mockFetch({ id: 1 });
    const result = await apiPost("accounts", { Username: "test" });
    expect(fetch).toHaveBeenCalledWith("/api/proxy/accounts", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ Username: "test" }),
    }));
    expect(result).toEqual({ id: 1 });
  });
});

describe("apiPatch", () => {
  it("sends PATCH with JSON body", async () => {
    mockFetch({ updated: true });
    await apiPatch("emails/1", { IsRead: true });
    expect(fetch).toHaveBeenCalledWith("/api/proxy/emails/1", expect.objectContaining({
      method: "PATCH",
      body: JSON.stringify({ IsRead: true }),
    }));
  });
});

describe("apiDelete", () => {
  it("sends DELETE and returns void", async () => {
    (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      status: 204,
    });
    const result = await apiDelete("emails/1");
    expect(fetch).toHaveBeenCalledWith("/api/proxy/emails/1", expect.objectContaining({
      method: "DELETE",
    }));
    expect(result).toBeUndefined();
  });
});

describe("error handling", () => {
  it("throws ApiError with status and message from body", async () => {
    mockFetch({ error: "Not found" }, 404);
    await expect(apiGet("missing")).rejects.toThrow(ApiError);
    try {
      mockFetch({ error: "Not found" }, 404);
      await apiGet("missing");
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).status).toBe(404);
      expect((err as ApiError).message).toBe("Not found");
    }
  });

  it("falls back to default message when body is unparseable", async () => {
    mockFetchJsonError(500);
    try {
      await apiGet("broken");
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).status).toBe(500);
      expect((err as ApiError).message).toBe("Request failed");
    }
  });
});
