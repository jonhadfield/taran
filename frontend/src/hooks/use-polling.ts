"use client";

import { useState, useEffect } from "react";
import { apiGet } from "@/lib/api";

export function usePolling<T>(
  path: string,
  initialData: T,
  intervalMs = 60_000
): T {
  const [data, setData] = useState(initialData);

  useEffect(() => {
    let cancelled = false;

    // Fetch immediately when path changes
    apiGet<T>(path)
      .then((res) => {
        if (!cancelled) setData(res);
      })
      .catch(() => {});

    const id = setInterval(async () => {
      try {
        const res = await apiGet<T>(path);
        if (!cancelled) setData(res);
      } catch {
        // Keep showing stale data on error
      }
    }, intervalMs);

    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [path, intervalMs]);

  return data;
}
