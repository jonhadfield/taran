"use client";

import { useState, useEffect, useRef } from "react";
import { apiGet } from "@/lib/api";

export function usePolling<T>(
  path: string,
  initialData: T,
  intervalMs = 15_000
): T {
  const [data, setData] = useState(initialData);
  const pathRef = useRef(path);
  pathRef.current = path;

  useEffect(() => {
    const id = setInterval(async () => {
      try {
        const res = await apiGet<T>(pathRef.current);
        setData(res);
      } catch {
        // Keep showing stale data on error
      }
    }, intervalMs);

    return () => clearInterval(id);
  }, [intervalMs]);

  return data;
}
