"use client";

import { usePolling } from "@/hooks/use-polling";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { DashboardData, WeekCount, TopicCount, CategoryCount, HeatmapCell } from "@/types/api";
import { Inbox, BookOpen, Mail, TrendingUp, TrendingDown, Loader2, AlertCircle, BarChart3, Tag, CheckSquare, PieChart, Clock } from "lucide-react";
import { CopyEmailAddress } from "@/components/copy-email-address";
import Link from "next/link";

interface DashboardContentProps {
  initialData: DashboardData;
  emailAddress: string;
}

function WeeklyChart({ data }: { data: WeekCount[] }) {
  if (data.length < 2) return null;

  const max = Math.max(...data.map((d) => d.Count), 1);

  return (
    <div className="flex items-end gap-1.5 h-24">
      {data.map((week, i) => {
        const height = Math.max(4, (week.Count / max) * 100);
        const label = new Date(week.Week).toLocaleDateString("en-US", {
          month: "short",
          day: "numeric",
        });
        return (
          <div
            key={i}
            className="flex-1 flex flex-col items-center gap-1 min-w-0"
          >
            <span className="text-[10px] text-muted-foreground tabular-nums">
              {week.Count > 0 ? week.Count : ""}
            </span>
            <div
              className="w-full rounded-sm bg-primary/80 transition-all hover:bg-primary"
              style={{ height: `${height}%` }}
              title={`${label}: ${week.Count} emails`}
            />
            <span className="text-[10px] text-muted-foreground truncate w-full text-center hidden @sm:block">
              {label}
            </span>
          </div>
        );
      })}
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

const CATEGORY_COLORS: Record<string, string> = {
  newsletter: "bg-blue-500",
  personal: "bg-green-500",
  transactional: "bg-gray-400",
  marketing: "bg-orange-500",
  notification: "bg-yellow-500",
  other: "bg-gray-300 dark:bg-gray-600",
};

function CategoryBars({ categories }: { categories: CategoryCount[] }) {
  if (categories.length === 0) return null;

  const max = Math.max(...categories.map((c) => c.Count), 1);

  return (
    <div className="space-y-2">
      {categories.map((cat) => (
        <div key={cat.Category} className="flex items-center gap-3">
          <span className="text-xs w-24 text-right truncate capitalize text-muted-foreground">
            {cat.Category || "other"}
          </span>
          <div className="flex-1 h-5 bg-muted rounded-sm overflow-hidden">
            <div
              className={`h-full rounded-sm transition-all ${CATEGORY_COLORS[cat.Category] || CATEGORY_COLORS.other}`}
              style={{ width: `${(cat.Count / max) * 100}%` }}
            />
          </div>
          <span className="text-xs tabular-nums text-muted-foreground w-8">
            {cat.Count}
          </span>
        </div>
      ))}
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
              className={`flex-1 aspect-square rounded-[2px] transition-colors ${getColor(count)}`}
              title={`${DAY_LABELS[day]} ${hour}:00 — ${count} email${count !== 1 ? "s" : ""}`}
            />
          ))}
        </div>
      ))}
    </div>
  );
}

export function DashboardContent({ initialData, emailAddress }: DashboardContentProps) {
  const { data } = usePolling<DashboardData>("dashboard", initialData);

  const emails = data.emails || [];
  const digests = data.digests || [];
  const unreadCount = data.unreadCount;
  const stats = data.stats;
  const weeklyHistory = data.weeklyHistory || [];
  const topicsWithCount = data.topicsWithCount || [];
  const categories = data.categories || [];
  const actionItems = data.actionItems || 0;
  const heatmap = data.heatmap || [];

  const processing = data.processing;
  const weekDiff = stats.EmailsThisWeek - stats.EmailsLastWeek;
  const inFlightCount = (processing?.PendingCount ?? 0) + (processing?.ProcessingCount ?? 0);
  const failedCount = processing?.FailedCount ?? 0;

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold text-balance">Dashboard</h1>

      {/* Stat cards */}
      <div className="grid gap-4 @sm:grid-cols-2 @lg:grid-cols-4">
        <Card className="hover:shadow-md transition-shadow">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex size-12 items-center justify-center rounded-lg bg-blue-500/10">
              <Mail className="size-6 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <p className="text-2xl font-bold">{stats.EmailsThisWeek}</p>
              <p className="text-sm text-muted-foreground">This Week</p>
              {stats.EmailsLastWeek > 0 && (
                <div className="flex items-center gap-1 mt-0.5">
                  {weekDiff >= 0 ? (
                    <TrendingUp className="size-3 text-emerald-500" />
                  ) : (
                    <TrendingDown className="size-3 text-rose-500" />
                  )}
                  <span className={`text-xs ${weekDiff >= 0 ? "text-emerald-600" : "text-rose-600"}`}>
                    {weekDiff >= 0 ? "+" : ""}{weekDiff} vs last week
                  </span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        <Card className="hover:shadow-md transition-shadow">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex size-12 items-center justify-center rounded-lg bg-amber-500/10">
              <Inbox className="size-6 text-amber-600 dark:text-amber-400" />
            </div>
            <div>
              <p className="text-2xl font-bold">{unreadCount}</p>
              <p className="text-sm text-muted-foreground">Unread</p>
            </div>
          </CardContent>
        </Card>

        <Card className="hover:shadow-md transition-shadow">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex size-12 items-center justify-center rounded-lg bg-indigo-500/10">
              <BookOpen className="size-6 text-indigo-600 dark:text-indigo-400" />
            </div>
            <div>
              <p className="text-2xl font-bold">{stats.TotalEmails}</p>
              <p className="text-sm text-muted-foreground">Total Emails</p>
            </div>
          </CardContent>
        </Card>

        <Card className="hover:shadow-md transition-shadow">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex size-12 items-center justify-center rounded-lg bg-emerald-500/10">
              <CheckSquare className="size-6 text-emerald-600 dark:text-emerald-400" />
            </div>
            <div>
              <p className="text-2xl font-bold">{actionItems}</p>
              <p className="text-sm text-muted-foreground">Action Items</p>
              <p className="text-xs text-muted-foreground">This week</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Processing status */}
      {(inFlightCount > 0 || failedCount > 0) && (
        <div className="flex flex-wrap items-center gap-4 text-sm">
          {inFlightCount > 0 && (
            <div className="flex items-center gap-1.5 text-muted-foreground">
              <Loader2 className="size-3.5 animate-spin" />
              <span>{inFlightCount} email{inFlightCount !== 1 ? "s" : ""} processing...</span>
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

      {/* Topic cloud */}
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
                  <span className="size-2 shrink-0 rounded-full bg-blue-500" />
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
                  {new Date(email.ReceivedAt).toLocaleDateString()}
                </span>
              </Link>
            ))}
          </div>
        </div>
      )}

      {/* Empty state */}
      {emails.length === 0 && digests.length === 0 && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <Mail className="size-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium">No emails yet</h3>
            <p className="text-sm text-muted-foreground mt-2">
              Forward your newsletters to your inbox to get started:
            </p>
            {emailAddress && (
              <div className="mt-3">
                <CopyEmailAddress emailAddress={emailAddress} />
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
