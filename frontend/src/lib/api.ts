export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`/api/proxy/${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Request failed" }));
    throw new ApiError(res.status, body.error || "Request failed");
  }

  return res.json();
}

export function apiGet<T>(path: string, params?: Record<string, string>): Promise<T> {
  const query = params ? "?" + new URLSearchParams(params).toString() : "";
  return request<T>(`${path}${query}`);
}

export function apiPost<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function apiPatch<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, {
    method: "PATCH",
    body: JSON.stringify(body),
  });
}

export function apiDeleteJSON<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, {
    method: "DELETE",
    body: JSON.stringify(body),
  });
}

export async function apiDelete(path: string): Promise<void> {
  const res = await fetch(`/api/proxy/${path}`, {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Request failed" }));
    throw new ApiError(res.status, body.error || "Request failed");
  }
}
