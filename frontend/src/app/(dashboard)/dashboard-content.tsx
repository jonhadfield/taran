"use client";

import { usePolling } from "@/hooks/use-polling";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { DashboardData } from "@/types/api";
import { Inbox, BookOpen, Mail, TrendingUp, TrendingDown } from "lucide-react";
import { APP_NAME } from "@/lib/config";
import Link from "next/link";

interface DashboardContentProps {
  initialData: DashboardData;
}

export function DashboardContent({ initialData }: DashboardContentProps) {
  const data = usePolling<DashboardData>("dashboard", initialData);

  const emails = data.emails || [];
  const digests = data.digests || [];
  const unreadCount = data.unreadCount;
  const stats = data.stats;

  const weekDiff = stats.EmailsThisWeek - stats.EmailsLastWeek;

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold text-balance">Dashboard</h1>

      {/* Stat cards */}
      <div className="grid gap-4 @sm:grid-cols-2 @lg:grid-cols-3">
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
      </div>

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
                  <div className="space-y-1">
                    <p className="font-medium">{digest.Title}</p>
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {digest.Summary}
                    </p>
                  </div>
                  <Badge variant="secondary">{digest.EmailCount} emails</Badge>
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
            <p className="text-sm text-muted-foreground mt-1">
              Forward your newsletters to your {APP_NAME} inbox to get started.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
