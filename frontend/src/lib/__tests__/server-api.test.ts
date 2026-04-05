import { describe, it, expect, vi, beforeEach } from "vitest";

const mockCookies = vi.fn();

vi.mock("next/headers", () => ({
  cookies: mockCookies,
}));

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
  vi.stubEnv("BACKEND_URL", "http://localhost:8080");
  vi.stubEnv("API_KEY", "test-api-key");
});

async function loadServerFetch() {
  vi.resetModules();
  const mod = await import("../server-api");
  return mod.serverFetch;
}

describe("serverFetch", () => {
  it("throws 'Not authenticated' when no session cookie exists", async () => {
    mockCookies.mockResolvedValue({
      get: () => undefined,
    });
    const serverFetch = await loadServerFetch();
    await expect(serverFetch("emails")).rejects.toThrow("Not authenticated");
  });

  it("strips HMAC signature from cookie value", async () => {
    mockCookies.mockResolvedValue({
      get: (name: string) =>
        name === "better-auth.session_token"
          ? { value: "token123.hmac_sig" }
          : undefined,
    });
    (fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: [] }),
    });
    const serverFetch = await loadServerFetch();
    await serverFetch("emails");

    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/emails",
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: "Bearer token123",
        }),
      }),
    );
  });

  it("sends correct auth headers", async () => {
    mockCookies.mockResolvedValue({
      get: (name: string) =>
        name === "better-auth.session_token"
          ? { value: "simple-token" }
          : undefined,
    });
    (fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: [] }),
    });
    const serverFetch = await loadServerFetch();
    await serverFetch("accounts");

    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/accounts",
      expect.objectContaining({
        headers: {
          Authorization: "Bearer simple-token",
          "X-API-Key": "test-api-key",
        },
        cache: "no-store",
      }),
    );
  });

  it("returns parsed JSON on success", async () => {
    mockCookies.mockResolvedValue({
      get: (name: string) =>
        name === "better-auth.session_token"
          ? { value: "token" }
          : undefined,
    });
    (fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ emails: [{ id: 1 }] }),
    });
    const serverFetch = await loadServerFetch();
    const result = await serverFetch("emails");
    expect(result).toEqual({ emails: [{ id: 1 }] });
  });

  it("throws on non-ok response", async () => {
    mockCookies.mockResolvedValue({
      get: (name: string) =>
        name === "better-auth.session_token"
          ? { value: "token" }
          : undefined,
    });
    (fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: false,
      status: 500,
    });
    const serverFetch = await loadServerFetch();
    await expect(serverFetch("emails")).rejects.toThrow("API error: 500");
  });
});
