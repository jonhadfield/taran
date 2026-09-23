import { describe, it, expect, vi, beforeEach } from "vitest";

const mockCookies = vi.fn();
const mockHeaders = vi.fn();

vi.mock("next/headers", () => ({
  cookies: mockCookies,
  headers: mockHeaders,
}));

beforeEach(() => {
  mockHeaders.mockResolvedValue(new Headers());
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

describe("serverFetch client IP", () => {
  beforeEach(() => {
    mockCookies.mockResolvedValue({
      get: (name: string) =>
        name === "better-auth.session_token" ? { value: "token123.sig" } : undefined,
    });
    (fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });
  });

  function sentHeaders() {
    const [, init] = (fetch as ReturnType<typeof vi.fn>).mock.calls[0];
    return init.headers as Record<string, string>;
  }

  it("forwards the visitor's address so the backend rate-limits per user", async () => {
    mockHeaders.mockResolvedValue(new Headers({ "x-forwarded-for": "203.0.113.9, 10.0.0.1" }));
    await (await loadServerFetch())("emails");
    expect(sentHeaders()["X-Client-IP"]).toBe("203.0.113.9");
  });

  it("omits the header when no address is available", async () => {
    await (await loadServerFetch())("emails");
    expect(sentHeaders()).not.toHaveProperty("X-Client-IP");
  });

  it("still sends the request if headers are unavailable", async () => {
    mockHeaders.mockRejectedValue(new Error("called outside a request scope"));
    await expect((await loadServerFetch())("emails")).resolves.toEqual({});
    expect(sentHeaders()).not.toHaveProperty("X-Client-IP");
  });
});
