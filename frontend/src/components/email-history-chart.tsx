"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { WeekCount } from "@/types/api";

export function EmailHistoryChart() {
  const [data, setData] = useState<WeekCount[]>([]);

  useEffect(() => {
    async function load() {
      try {
        const res = await fetch("/api/proxy/stats/history?weeks=8");
        if (!res.ok) return;
        const json = await res.json();
        if (Array.isArray(json)) setData(json);
      } catch {
        // Silently ignore
      }
    }
    load();
  }, []);

  if (data.length === 0) return null;

  const maxCount = Math.max(...data.map((d) => d.Count), 1);
  const barMaxHeight = 120;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Emails Per Week</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-end gap-2" style={{ height: barMaxHeight + 28 }}>
          {data.map((week) => {
            const height = Math.max((week.Count / maxCount) * barMaxHeight, 4);
            const weekDate = new Date(week.Week);
            const label = `${weekDate.getMonth() + 1}/${weekDate.getDate()}`;
            return (
              <div
                key={week.Week}
                className="flex flex-1 flex-col items-center gap-1"
              >
                <span className="text-xs text-muted-foreground">
                  {week.Count}
                </span>
                <div
                  className="w-full rounded-t bg-blue-500/80 transition-all"
                  style={{ height }}
                />
                <span className="text-xs text-muted-foreground">{label}</span>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
