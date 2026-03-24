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

function getBarColor(percent: number): string {
  if (percent >= 90) return "bg-red-500";
  if (percent >= 70) return "bg-yellow-500";
  return "bg-primary";
}

function getTextColor(percent: number): string {
  if (percent >= 90) return "text-red-600 dark:text-red-400";
  if (percent >= 70) return "text-yellow-600 dark:text-yellow-400";
  return "";
}

export function UsageStatsCard() {
  const [stats, setStats] = useState<UsageStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [now] = useState(() => Date.now());
  const [todayStr] = useState(() => new Date().toISOString().slice(0, 10));

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

  const hasMonthlyLimit = stats.MonthlyTokenLimit > 0;
  const hasDailyLimit = stats.DailyTokenLimit > 0;
  const monthlyPercent = hasMonthlyLimit
    ? Math.min(100, (stats.MonthlyTokensUsed / stats.MonthlyTokenLimit) * 100)
    : 0;
  const dailyPercent = hasDailyLimit
    ? Math.min(100, (stats.DailyTokensUsed / stats.DailyTokenLimit) * 100)
    : 0;

  const periodLabel = stats.PeriodStart
    ? new Date(stats.PeriodStart).toLocaleDateString("en-GB", {
        month: "long",
        year: "numeric",
      })
    : "This month";

  // Calculate projected monthly usage based on daily average
  const history = stats.DailyHistory || [];
  const daysInMonth = stats.PeriodEnd && stats.PeriodStart
    ? Math.round((new Date(stats.PeriodEnd).getTime() - new Date(stats.PeriodStart).getTime()) / (1000 * 60 * 60 * 24))
    : 30;
  const daysSoFar = stats.PeriodStart
    ? Math.max(1, Math.round((now - new Date(stats.PeriodStart).getTime()) / (1000 * 60 * 60 * 24)))
    : 1;
  const projectedMonthly = Math.round((stats.MonthlyTokensUsed / daysSoFar) * daysInMonth);
  const projectedPercent = hasMonthlyLimit
    ? Math.min(100, (projectedMonthly / stats.MonthlyTokenLimit) * 100)
    : 0;

  // Find max for history chart scaling
  const maxHistoryTokens = history.length > 0
    ? Math.max(...history.map((d) => d.Tokens), 1)
    : 1;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Usage</CardTitle>
        <CardDescription>
          AI token usage for {periodLabel}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Monthly usage */}
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Monthly tokens</span>
            <span className={`font-medium ${hasMonthlyLimit ? getTextColor(monthlyPercent) : ""}`}>
              {formatTokens(stats.MonthlyTokensUsed)}
              {hasMonthlyLimit && ` / ${formatTokens(stats.MonthlyTokenLimit)}`}
            </span>
          </div>
          {hasMonthlyLimit && (
            <>
              <div className="h-2 rounded-full bg-muted overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all ${getBarColor(monthlyPercent)}`}
                  style={{ width: `${monthlyPercent}%` }}
                />
              </div>
              {monthlyPercent >= 80 && (
                <p className="text-xs text-red-600 dark:text-red-400 font-medium">
                  {monthlyPercent >= 100
                    ? "Monthly limit reached — email processing is paused until next month"
                    : `${Math.round(monthlyPercent)}% of monthly limit used`}
                </p>
              )}
            </>
          )}
        </div>

        {/* Projected usage */}
        {hasMonthlyLimit && daysSoFar > 1 && (
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Projected monthly</span>
            <span className={`font-medium ${getTextColor(projectedPercent)}`}>
              {formatTokens(projectedMonthly)}
              {projectedPercent > 100 && " (over limit)"}
            </span>
          </div>
        )}

        {/* Daily usage */}
        <div className="space-y-2 pt-2 border-t">
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Today</span>
            <span className={`font-medium ${hasDailyLimit ? getTextColor(dailyPercent) : ""}`}>
              {formatTokens(stats.DailyTokensUsed)}
              {hasDailyLimit && ` / ${formatTokens(stats.DailyTokenLimit)}`}
            </span>
          </div>
          {hasDailyLimit && (
            <div className="h-1.5 rounded-full bg-muted overflow-hidden">
              <div
                className={`h-full rounded-full transition-all ${getBarColor(dailyPercent)}`}
                style={{ width: `${dailyPercent}%` }}
              />
            </div>
          )}
        </div>

        {/* Operation breakdown */}
        <div className="grid grid-cols-3 gap-4 pt-2 border-t">
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

        {/* 30-day history chart */}
        {history.length > 0 && (
          <div className="space-y-2 pt-2 border-t">
            <div className="text-sm text-muted-foreground">Last 30 days</div>
            <div
              className="grid gap-px"
              style={{ gridTemplateColumns: `repeat(${history.length}, 1fr)`, height: 64 }}
            >
              {history.map((day) => {
                const pct = Math.max(2, (day.Tokens / maxHistoryTokens) * 100);
                const isToday = day.Date === todayStr;
                return (
                  <div
                    key={day.Date}
                    className="flex flex-col justify-end group/bar relative"
                    title={`${day.Date}: ${formatTokens(day.Tokens)} tokens`}
                  >
                    <div
                      className={`w-full rounded-sm transition-colors ${
                        isToday ? "bg-primary" : "bg-muted-foreground/30 hover:bg-muted-foreground/50"
                      }`}
                      style={{ height: `${pct}%` }}
                    />
                  </div>
                );
              })}
            </div>
            <div className="flex justify-between text-[10px] text-muted-foreground">
              <span>{history[0]?.Date.slice(5)}</span>
              <span>Today</span>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
