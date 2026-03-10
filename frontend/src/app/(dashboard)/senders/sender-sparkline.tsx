"use client";

import { useState, useEffect } from "react";
import { apiGet } from "@/lib/api";
import type { WeekCount } from "@/types/api";

function Sparkline({
  data,
  width = 120,
  height = 24,
}: {
  data: number[];
  width?: number;
  height?: number;
}) {
  if (data.length < 2) return null;
  const max = Math.max(...data, 1);
  const points = data
    .map(
      (v, i) =>
        `${(i / (data.length - 1)) * width},${height - (v / max) * (height - 2) - 1}`
    )
    .join(" ");

  return (
    <svg width={width} height={height} className="text-muted-foreground">
      <polyline
        points={points}
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function SenderSparkline({ fromAddress }: { fromAddress: string }) {
  const [data, setData] = useState<number[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    apiGet<WeekCount[]>("senders/history", {
      address: fromAddress,
      weeks: "12",
    })
      .then((counts) => {
        if (cancelled) return;
        setData((counts || []).map((wc) => wc.Count));
      })
      .catch(() => {
        if (!cancelled) setData([]);
      });
    return () => {
      cancelled = true;
    };
  }, [fromAddress]);

  if (data === null) return null;
  if (data.length < 2) return null;

  return <Sparkline data={data} />;
}
