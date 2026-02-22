"use client";

import { useState, useEffect } from "react";
import { apiGet } from "@/lib/api";

interface PollingResult<T> {
  data: T;
  loading: boolean;
  error: string | null;
}

export function usePolling<T>(
  path: string,
  initialData: T,
  intervalMs = 60_000
): PollingResult<T> {
  const [data, setData] = useState(initialData);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    // Fetch immediately when path changes
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
  }, [path, intervalMs]);

  return { data, loading, error };
}
