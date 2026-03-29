"use client";

import { useState, useEffect, useCallback } from "react";
import { apiGet } from "@/lib/api";

interface PollingResult<T> {
  data: T;
  loading: boolean;
  error: string | null;
  refresh: () => void;
}

export function usePolling<T>(
  path: string,
  initialData: T,
  intervalMs = 60_000
): PollingResult<T> {
  const [data, setData] = useState(initialData);
  // Start as not-loading when initialData is provided (data is already available)
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [refreshCount, setRefreshCount] = useState(0);

  const refresh = useCallback(() => setRefreshCount((c) => c + 1), []);

  useEffect(() => {
    let cancelled = false;

    // Fetch immediately when path changes or refresh is called
    apiGet<T>(path)
      .then((res) => {
        if (!cancelled) {
          setData(res);
          setError(null);
          setLoading(false);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Fetch failed");
          setLoading(false);
        }
      });

    const id = setInterval(async () => {
      try {
        const res = await apiGet<T>(path);
        if (!cancelled) {
          setData(res);
          setError(null);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Fetch failed");
        }
      }
    }, intervalMs);

    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [path, intervalMs, refreshCount]);

  return { data, loading, error, refresh };
}
