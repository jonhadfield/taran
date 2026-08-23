"use client";

import { useState, useEffect } from "react";
import { usePolling } from "@/hooks/use-polling";
import { apiGet } from "@/lib/api";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { DashboardData, UserPreference, WeekCount, TopicCount, CategoryCount, HeatmapCell } from "@/types/api";
import { Inbox, BookOpen, Mail, TrendingUp, TrendingDown, AlertCircle, BarChart3, Tag, PieChart, Clock } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import { CopyEmailAddress } from "@/components/copy-email-address";
import { OnboardingChecklist } from "@/components/onboarding-checklist";
import { FaviconBadge } from "@/components/favicon-badge";
import { DashboardIllustration } from "@/components/empty-state";
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty";
import { WidgetErrorBoundary } from "@/components/widget-error-boundary";
import Link from "next/link";
import { formatShortDate, pluralize } from "@/lib/utils";
import { CATEGORY_CHART_COLORS } from "@/lib/category-constants";

interface DashboardContentProps {
  initialData: DashboardData;
  emailAddress: string;
}

function WeeklyChart({ data }: { data: WeekCount[] }) {
  if (data.length < 2) return null;

  const max = Math.max(...data.map((d) => d.Count), 1);

  return (
    <div className="space-y-1">
      <div
        className="grid gap-1.5"
        style={{ gridTemplateColumns: `repeat(${data.length}, 1fr)`, height: 96 }}
      >
        {data.map((week, i) => {
          const pct = Math.max(4, (week.Count / max) * 100);
          const label = formatShortDate(week.Week);
          return (
            <div key={i} className="group relative flex flex-col items-center justify-end gap-1 min-w-0">
              <div className="absolute -top-7 left-1/2 -translate-x-1/2 hidden group-hover:block z-10 whitespace-nowrap rounded bg-popover border px-2 py-1 text-xs shadow-md">
                {label}: <span className="font-medium">{week.Count}</span> {pluralize(week.Count, "email")}
              </div>
              <span className="text-[10px] text-muted-foreground tabular-nums">
                {week.Count > 0 ? week.Count : ""}
              </span>
              <div
                className="w-full rounded-sm bg-primary/80 hover:bg-primary animate-bar-grow"
                style={{ "--bar-height": `${pct}%`, animationDelay: `${i * 60}ms` } as React.CSSProperties}
              />
            </div>
          );
        })}
      </div>
      <div
        className="grid gap-1.5"
        style={{ gridTemplateColumns: `repeat(${data.length}, 1fr)` }}
      >
        {data.map((week, i) => {
          const label = formatShortDate(week.Week);
          return (
            <span key={i} className="text-[10px] text-muted-foreground truncate text-center hidden @sm:block">
              {label}
            </span>
          );
        })}
      </div>
    </div>
  );
}

function TopicCloud({ topics }: { topics: TopicCount[] }) {
  if (topics.length === 0) return null;

  return (
    <div className="flex flex-wrap gap-2">
      {topics.map((tc) => (
        <Link
          key={tc.Topic}
          href={`/inbox?topic=${encodeURIComponent(tc.Topic)}`}
          className="inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
        >
          {tc.Topic}
          <span className="text-muted-foreground tabular-nums">{tc.Count}</span>
        </Link>
      ))}
    </div>
  );
}


function CategoryBars({ categories }: { categories: CategoryCount[] }) {
  if (categories.length === 0) return null;

  const max = Math.max(...categories.map((c) => c.Count), 1);
  const total = categories.reduce((s, c) => s + c.Count, 0);

  return (
    <div className="space-y-2">
      {categories.map((cat, i) => {
        const pct = ((cat.Count / total) * 100).toFixed(1);
        return (
          <div key={cat.Category} className="group relative flex items-center gap-3">
            <div className="absolute -top-7 left-28 hidden group-hover:block z-10 whitespace-nowrap rounded bg-popover border px-2 py-1 text-xs shadow-md">
              <span className="capitalize font-medium">{cat.Category || "other"}</span>: {cat.Count} ({pct}%)
            </div>
            <span className="text-xs w-24 text-right truncate capitalize text-muted-foreground">
              {cat.Category || "other"}
            </span>
            <div className="flex-1 h-5 bg-muted rounded-sm overflow-hidden">
              <div
                className={`h-full rounded-sm animate-bar-grow-h ${CATEGORY_CHART_COLORS[cat.Category] || CATEGORY_CHART_COLORS.other}`}
                style={{ "--bar-width": `${(cat.Count / max) * 100}%`, animationDelay: `${i * 80}ms` } as React.CSSProperties}
              />
            </div>
            <span className="text-xs tabular-nums text-muted-foreground w-8">
              {cat.Count}
            </span>
          </div>
        );
      })}
    </div>
  );
}

const DAY_LABELS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

function ArrivalHeatmap({ cells }: { cells: HeatmapCell[] }) {
  if (cells.length === 0) return null;

  // Build a 7×24 grid (day × hour)
  const grid: number[][] = Array.from({ length: 7 }, () => Array(24).fill(0));
  let max = 0;
  for (const cell of cells) {
    grid[cell.Day][cell.Hour] = cell.Count;
    if (cell.Count > max) max = cell.Count;
  }

  const getColor = (count: number) => {
    if (count === 0) return "bg-muted";
    const intensity = count / max;
    if (intensity < 0.25) return "bg-primary/20";
    if (intensity < 0.5) return "bg-primary/40";
    if (intensity < 0.75) return "bg-primary/60";
    return "bg-primary/90";
  };

  // Show hours in 3-hour increments for labels
  const hourLabels = [0, 3, 6, 9, 12, 15, 18, 21];

  return (
    <div className="space-y-1">
      {/* Hour labels */}
      <div className="flex items-center gap-px ml-9">
        {Array.from({ length: 24 }, (_, h) => (
          <div key={h} className="flex-1 text-center">
            {hourLabels.includes(h) && (
              <span className="text-[9px] text-muted-foreground tabular-nums">
                {h === 0 ? "12a" : h < 12 ? `${h}a` : h === 12 ? "12p" : `${h - 12}p`}
              </span>
            )}
          </div>
        ))}
      </div>
      {/* Grid rows */}
      {grid.map((row, day) => (
        <div key={day} className="flex items-center gap-px">
          <span className="text-[10px] text-muted-foreground w-8 text-right pr-1 shrink-0">
            {DAY_LABELS[day]}
          </span>
          {row.map((count, hour) => (
            <div
              key={hour}
              className={`group relative flex-1 aspect-square rounded-[2px] transition-colors ${getColor(count)}`}
            >
              {count > 0 && (
                <div className="absolute -top-8 left-1/2 -translate-x-1/2 hidden group-hover:block z-10 whitespace-nowrap rounded bg-popover border px-2 py-1 text-xs shadow-md">
                  {DAY_LABELS[day]} {hour}:00 — <span className="font-medium">{count}</span>
                </div>
              )}
            </div>
          ))}
        </div>
      ))}
    </div>
  );
}

export function DashboardContent({ initialData, emailAddress }: DashboardContentProps) {
  const { data } = usePolling<DashboardData>("dashboard", initialData);
  const [hasConfiguredPrefs, setHasConfiguredPrefs] = useState(false);

  useEffect(() => {
    apiGet<UserPreference>("preferences")
      .then((pref) => {
        // Consider preferences "configured" if the user has enabled any delivery
        // method or changed the timezone from UTC (the default)
        if (pref.DigestEmail || pref.DigestWebhook || pref.WeeklySummary === false || pref.DigestTimezone !== "UTC") {
          setHasConfiguredPrefs(true);
        }
      })
      .catch(() => {});
  }, []);

  const emails = data.Emails || [];
  const digests = data.Digests || [];
  const unreadCount = data.UnreadCount;
  const stats = data.Stats;
  const weeklyHistory = data.WeeklyHistory || [];
  const topicsWithCount = data.TopicsWithCount || [];
  const categories = data.Categories || [];
  const heatmap = data.Heatmap || [];

  const processing = data.Processing;
  const weekDiff = stats.EmailsThisWeek - stats.EmailsLastWeek;
  const inFlightCount = (processing?.PendingCount ?? 0) + (processing?.ProcessingCount ?? 0);
  const failedCount = processing?.FailedCount ?? 0;

  return (
    <div className="space-y-8">
      <FaviconBadge count={unreadCount} />
      <h1 className="text-2xl font-bold text-balance">Dashboard</h1>

      {/* Stat cards */}
      <div className="grid gap-4 @sm:grid-cols-2 @lg:grid-cols-3">
        <Card className="hover:shadow-md transition-shadow">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex size-12 items-center justify-center rounded-lg bg-chart-1/10">
              <Mail className="size-6 text-chart-1" />
            </div>
            <div>
              <p className="text-2xl font-bold">{stats.EmailsThisWeek}</p>
              <p className="text-sm text-muted-foreground">This Week</p>
              {stats.EmailsLastWeek > 0 && (
                <div className="flex items-center gap-1 mt-0.5">
                  {weekDiff >= 0 ? (
                    <TrendingUp className="size-3 text-success" />
                  ) : (
                    <TrendingDown className="size-3 text-destructive" />
                  )}
                  <span className={`text-xs ${weekDiff >= 0 ? "text-success" : "text-destructive"}`}>
                    {weekDiff >= 0 ? "+" : ""}{weekDiff} vs last week
                  </span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        <Card className="hover:shadow-md transition-shadow">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex size-12 items-center justify-center rounded-lg bg-chart-2/10">
              <Inbox className="size-6 text-chart-2" />
            </div>
            <div>
              <p className="text-2xl font-bold">{unreadCount}</p>
              <p className="text-sm text-muted-foreground">Unread</p>
            </div>
          </CardContent>
        </Card>

        <Card className="hover:shadow-md transition-shadow">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex size-12 items-center justify-center rounded-lg bg-chart-3/10">
              <BookOpen className="size-6 text-chart-3" />
            </div>
            <div>
              <p className="text-2xl font-bold">{stats.TotalEmails}</p>
              <p className="text-sm text-muted-foreground">Total Emails</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Onboarding checklist — persistent until dismissed */}
      <OnboardingChecklist
        emailAddress={emailAddress}
        hasEmails={stats.TotalEmails > 0}
        hasDigest={digests.length > 0}
        hasConfiguredPreferences={hasConfiguredPrefs}
      />

      {/* Processing status */}
      {(inFlightCount > 0 || failedCount > 0) && (
        <div className="flex flex-wrap items-center gap-4 text-sm">
          {inFlightCount > 0 && (
            <div className="flex items-center gap-1.5 text-muted-foreground">
              <Spinner className="size-3.5" />
              <span>{inFlightCount} {pluralize(inFlightCount, "email")} processing...</span>
            </div>
          )}
          {failedCount > 0 && (
            <div className="flex items-center gap-1.5 text-muted-foreground">
              <AlertCircle className="size-3.5" />
              <span>{failedCount} failed</span>
            </div>
          )}
        </div>
      )}

      {/* Analytics row: Trend chart + Category distribution */}
      <WidgetErrorBoundary>
      {(weeklyHistory.length >= 2 || categories.length > 0) && (
        <div className="grid gap-4 @lg:grid-cols-2">
          {weeklyHistory.length >= 2 && (
            <Card>
              <CardContent className="p-5">
                <div className="flex items-center gap-2 mb-4">
                  <BarChart3 className="size-4 text-muted-foreground" />
                  <h3 className="text-sm font-semibold">Email Volume</h3>
                  <span className="text-xs text-muted-foreground ml-auto">Last 8 weeks</span>
                </div>
                <WeeklyChart data={weeklyHistory} />
              </CardContent>
            </Card>
          )}

          {categories.length > 0 && (
            <Card>
              <CardContent className="p-5">
                <div className="flex items-center gap-2 mb-4">
                  <PieChart className="size-4 text-muted-foreground" />
                  <h3 className="text-sm font-semibold">Categories</h3>
                </div>
                <CategoryBars categories={categories} />
              </CardContent>
            </Card>
          )}
        </div>
      )}
      </WidgetErrorBoundary>

      {/* Topic cloud + heatmap — wrapped so a broken chart doesn't crash the dashboard */}
      <WidgetErrorBoundary>
      {topicsWithCount.length > 0 && (
        <Card>
          <CardContent className="p-5">
            <div className="flex items-center gap-2 mb-4">
              <Tag className="size-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold">Top Topics</h3>
            </div>
            <TopicCloud topics={topicsWithCount} />
          </CardContent>
        </Card>
      )}

      {/* Arrival time heatmap */}
      {heatmap.length > 0 && (
        <Card>
          <CardContent className="p-5">
            <div className="flex items-center gap-2 mb-4">
              <Clock className="size-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold">Email Arrival Times</h3>
            </div>
            <ArrivalHeatmap cells={heatmap} />
          </CardContent>
        </Card>
      )}
      </WidgetErrorBoundary>

      {/* Top Senders */}
      {stats.TopSenders && stats.TopSenders.length > 0 && (
        <div className="space-y-3">
          <h2 className="text-lg font-semibold">Top Senders This Week</h2>
          <div className="divide-y rounded-lg border">
            {stats.TopSenders.map((sender) => (
              <div
                key={sender.FromAddress}
                className="flex items-center justify-between gap-3 px-4 py-2.5"
              >
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">
                    {sender.FromName || sender.FromAddress}
                  </p>
                  {sender.FromName && (
                    <p className="text-xs text-muted-foreground truncate">
                      {sender.FromAddress}
                    </p>
                  )}
                </div>
                <Badge variant="secondary" className="shrink-0 text-xs">
                  {sender.Count}
                </Badge>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Latest Digests */}
      {digests.length > 0 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Latest Digests</h2>
            <Link
              href="/digests"
              className="text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              View all
            </Link>
          </div>
          <div className="space-y-3">
            {digests.map((digest) => (
              <Link
                key={digest.ID}
                href={`/digests/${digest.ID}`}
                className="block rounded-lg border p-4 transition-all hover:shadow-md hover:bg-accent/50"
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="space-y-1 min-w-0">
                    <p className="font-medium truncate">{digest.Title}</p>
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {digest.Summary}
                    </p>
                  </div>
                  <Badge variant="secondary" className="shrink-0">{digest.EmailCount} emails</Badge>
                </div>
              </Link>
            ))}
          </div>
        </div>
      )}

      {/* Recent Emails */}
      {emails.length > 0 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Recent Emails</h2>
            <Link
              href="/inbox"
              className="text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              View all
            </Link>
          </div>
          <div className="divide-y rounded-lg border">
            {emails.map((email) => (
              <Link
                key={email.ID}
                href={`/inbox/${email.ID}`}
                className="flex items-center gap-3 p-3.5 transition-colors hover:bg-accent/50"
              >
                {!email.IsRead && (
                  <span className="size-2 shrink-0 rounded-full bg-info" />
                )}
                {email.IsRead && <span className="size-2 shrink-0" />}
                <div className="flex-1 min-w-0">
                  <p
                    className={`text-sm truncate ${!email.IsRead ? "font-semibold" : ""}`}
                  >
                    {email.FromName || email.FromAddress}
                  </p>
                  <p className="text-sm text-muted-foreground truncate">
                    {email.Subject}
                  </p>
                </div>
                <span className="text-xs text-muted-foreground whitespace-nowrap">
                  {formatShortDate(email.ReceivedAt)}
                </span>
              </Link>
            ))}
          </div>
        </div>
      )}

      {/* Empty state */}
      {emails.length === 0 && digests.length === 0 && (
        <Card>
          <CardContent>
            <Empty>
              <EmptyHeader>
                <EmptyMedia><DashboardIllustration /></EmptyMedia>
                <EmptyTitle>No emails yet</EmptyTitle>
                <EmptyDescription>Forward your newsletters to your inbox to get started:</EmptyDescription>
              </EmptyHeader>
              {emailAddress && (
                <EmptyContent>
                  <CopyEmailAddress emailAddress={emailAddress} />
                </EmptyContent>
              )}
            </Empty>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
