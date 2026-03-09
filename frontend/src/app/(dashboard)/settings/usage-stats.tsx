"use client";

import { useEffect, useState } from "react";
import { apiGet } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { UsageStats } from "@/types/api";

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return n.toString();
}

export function UsageStatsCard() {
  const [stats, setStats] = useState<UsageStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiGet<UsageStats>("usage")
      .then(setStats)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Usage</CardTitle>
          <CardDescription>Loading usage data...</CardDescription>
        </CardHeader>
      </Card>
    );
  }

  if (!stats) return null;

  const hasLimit = stats.MonthlyTokenLimit > 0;
  const usagePercent = hasLimit
    ? Math.min(100, (stats.MonthlyTokensUsed / stats.MonthlyTokenLimit) * 100)
    : 0;

  const periodLabel = stats.PeriodStart
    ? new Date(stats.PeriodStart).toLocaleDateString("en-US", {
        month: "long",
        year: "numeric",
      })
    : "This month";

  return (
    <Card>
      <CardHeader>
        <CardTitle>Usage</CardTitle>
        <CardDescription>
          AI token usage for {periodLabel}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Monthly tokens</span>
            <span className="font-medium">
              {formatTokens(stats.MonthlyTokensUsed)}
              {hasLimit && ` / ${formatTokens(stats.MonthlyTokenLimit)}`}
            </span>
          </div>
          {hasLimit && (
            <div className="h-2 rounded-full bg-muted overflow-hidden">
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${usagePercent}%` }}
              />
            </div>
          )}
        </div>

        <div className="grid grid-cols-3 gap-4 pt-2">
          <div className="text-center">
            <div className="text-lg font-semibold">
              {formatTokens(stats.TriageTokens)}
            </div>
            <div className="text-xs text-muted-foreground">Triage</div>
          </div>
          <div className="text-center">
            <div className="text-lg font-semibold">
              {formatTokens(stats.ExtractTokens)}
            </div>
            <div className="text-xs text-muted-foreground">Extraction</div>
          </div>
          <div className="text-center">
            <div className="text-lg font-semibold">
              {formatTokens(stats.DigestTokens)}
            </div>
            <div className="text-xs text-muted-foreground">Digest</div>
          </div>
        </div>

        <div className="flex justify-between text-sm pt-2 border-t">
          <span className="text-muted-foreground">Today</span>
          <span className="font-medium">
            {formatTokens(stats.DailyTokensUsed)} tokens
          </span>
        </div>
      </CardContent>
    </Card>
  );
}
