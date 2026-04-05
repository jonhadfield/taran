"use client";

import { useState } from "react";
import { toast } from "sonner";
import { usePolling } from "@/hooks/use-polling";
import { apiPatch } from "@/lib/api";
import { formatShortDate, formatTokens } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import type { AdminStats } from "@/types/api";
import { Users, Mail, BookOpen, TrendingUp, Cpu, Gauge, CheckCircle, XCircle, AlertTriangle, Clock } from "lucide-react";

function formatWeek(dateStr: string): string {
  const d = new Date(dateStr);
  return formatShortDate(d);
}

// MiniBarChart is an alias for the shared BarChart component
import { BarChart as MiniBarChart } from "@/components/bar-chart";

const emptyStats: AdminStats = {
  TotalUsers: 0,
  ActiveUsersWeek: 0,
  TotalEmails: 0,
  EmailsThisWeek: 0,
  TotalDigests: 0,
  DigestsThisWeek: 0,
  TopGlobalSenders: [],
  LLMProvider: "",
  LLMModel: "",
  MonthlyTokensUsed: 0,
  DefaultMonthlyTokenLimit: 0,
  ProcessedCount: 0,
  FailedCount: 0,
  SkippedCount: 0,
  PendingCount: 0,
  FeedbackUseful: 0,
  FeedbackNotUseful: 0,
  WeeklyEmails: [],
  WeeklyDigests: [],
  WeeklyTokens: [],
};

export function AdminDashboard() {
  const { data: stats, refresh } = usePolling<AdminStats>("admin/stats", emptyStats, 30_000);
  const [editingLimit, setEditingLimit] = useState(false);
  const [limitValue, setLimitValue] = useState("");
  const [saving, setSaving] = useState(false);

  const cards = [
    {
      title: "Total Users",
      value: stats.TotalUsers,
      sub: `${stats.ActiveUsersWeek} active this week`,
      icon: Users,
    },
    {
      title: "Total Emails",
      value: stats.TotalEmails,
      sub: `${stats.EmailsThisWeek} this week`,
      icon: Mail,
    },
    {
      title: "Total Digests",
      value: stats.TotalDigests,
      sub: `${stats.DigestsThisWeek} this week`,
      icon: BookOpen,
    },
    {
      title: "Active This Week",
      value: stats.ActiveUsersWeek,
      sub: `of ${stats.TotalUsers} total users`,
      icon: TrendingUp,
    },
  ];

  const totalProcessingEmails = stats.ProcessedCount + stats.FailedCount + stats.SkippedCount + stats.PendingCount;
  const successRate = totalProcessingEmails > 0
    ? ((stats.ProcessedCount / totalProcessingEmails) * 100).toFixed(1)
    : "0";

  const totalFeedback = stats.FeedbackUseful + stats.FeedbackNotUseful;
  const usefulRate = totalFeedback > 0
    ? ((stats.FeedbackUseful / totalFeedback) * 100).toFixed(0)
    : "—";

  const handleSaveLimit = async () => {
    const parsed = parseInt(limitValue, 10);
    if (isNaN(parsed) || parsed < 0) {
      toast.error("Enter a valid number (0 for unlimited)");
      return;
    }
    setSaving(true);
    try {
      await apiPatch("admin/settings/token-limit", { DefaultMonthlyTokenLimit: parsed });
      toast.success("Default token limit updated");
      setEditingLimit(false);
      refresh();
    } catch {
      toast.error("Failed to update token limit");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Summary cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {cards.map((card) => (
          <Card key={card.title}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {card.title}
              </CardTitle>
              <card.icon className="size-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{card.value.toLocaleString()}</div>
              <p className="text-xs text-muted-foreground mt-1">{card.sub}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Processing pipeline & feedback */}
      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Processing Pipeline
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="text-2xl font-bold">{successRate}%</div>
            <p className="text-xs text-muted-foreground -mt-2">success rate</p>
            <div className="grid grid-cols-2 gap-3 pt-2 border-t">
              <div className="flex items-center gap-2">
                <CheckCircle className="size-4 text-green-500" />
                <div>
                  <div className="text-sm font-medium">{stats.ProcessedCount.toLocaleString()}</div>
                  <div className="text-xs text-muted-foreground">Processed</div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <XCircle className="size-4 text-red-500" />
                <div>
                  <div className="text-sm font-medium">{stats.FailedCount.toLocaleString()}</div>
                  <div className="text-xs text-muted-foreground">Failed</div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <AlertTriangle className="size-4 text-yellow-500" />
                <div>
                  <div className="text-sm font-medium">{stats.SkippedCount.toLocaleString()}</div>
                  <div className="text-xs text-muted-foreground">Skipped</div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Clock className="size-4 text-blue-500" />
                <div>
                  <div className="text-sm font-medium">{stats.PendingCount.toLocaleString()}</div>
                  <div className="text-xs text-muted-foreground">Pending</div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              User Feedback
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="text-2xl font-bold">{usefulRate}{usefulRate !== "—" && "%"}</div>
            <p className="text-xs text-muted-foreground -mt-2">
              {totalFeedback > 0 ? `useful rate (${totalFeedback} ratings)` : "no feedback yet"}
            </p>
            {totalFeedback > 0 && (
              <div className="pt-2 border-t space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Useful</span>
                  <span className="font-medium text-green-600 dark:text-green-400">{stats.FeedbackUseful}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Not useful</span>
                  <span className="font-medium text-red-600 dark:text-red-400">{stats.FeedbackNotUseful}</span>
                </div>
                <div className="h-2 rounded-full bg-muted overflow-hidden">
                  <div
                    className="h-full rounded-full bg-green-500"
                    style={{ width: `${(stats.FeedbackUseful / totalFeedback) * 100}%` }}
                  />
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Weekly trends */}
      <div className="grid gap-4 sm:grid-cols-3">
        {stats.WeeklyEmails.length > 0 && (
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Emails / Week
              </CardTitle>
            </CardHeader>
            <CardContent>
              <MiniBarChart
                data={stats.WeeklyEmails.map((w) => ({
                  label: formatWeek(w.Week),
                  value: w.Count,
                }))}
              />
              <div className="flex justify-between text-[10px] text-muted-foreground mt-1">
                <span>{formatWeek(stats.WeeklyEmails[0].Week)}</span>
                <span>{formatWeek(stats.WeeklyEmails[stats.WeeklyEmails.length - 1].Week)}</span>
              </div>
            </CardContent>
          </Card>
        )}

        {stats.WeeklyDigests.length > 0 && (
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Digests / Week
              </CardTitle>
            </CardHeader>
            <CardContent>
              <MiniBarChart
                data={stats.WeeklyDigests.map((w) => ({
                  label: formatWeek(w.Week),
                  value: w.Count,
                }))}
              />
              <div className="flex justify-between text-[10px] text-muted-foreground mt-1">
                <span>{formatWeek(stats.WeeklyDigests[0].Week)}</span>
                <span>{formatWeek(stats.WeeklyDigests[stats.WeeklyDigests.length - 1].Week)}</span>
              </div>
            </CardContent>
          </Card>
        )}

        {stats.WeeklyTokens.length > 0 && (
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Tokens / Week
              </CardTitle>
            </CardHeader>
            <CardContent>
              <MiniBarChart
                data={stats.WeeklyTokens.map((w) => ({
                  label: formatWeek(w.Week),
                  value: w.Count,
                }))}
                formatValue={formatTokens}
              />
              <div className="flex justify-between text-[10px] text-muted-foreground mt-1">
                <span>{formatWeek(stats.WeeklyTokens[0].Week)}</span>
                <span>{formatWeek(stats.WeeklyTokens[stats.WeeklyTokens.length - 1].Week)}</span>
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      {/* LLM & Token config */}
      <div className="grid gap-4 sm:grid-cols-2">
        {stats.LLMProvider && (
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                LLM Provider
              </CardTitle>
              <Cpu className="size-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold capitalize">{stats.LLMProvider}</div>
              <p className="text-xs text-muted-foreground mt-1">
                Model: <span className="font-mono">{stats.LLMModel}</span>
              </p>
            </CardContent>
          </Card>
        )}

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Monthly Token Usage
            </CardTitle>
            <Gauge className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatTokens(stats.MonthlyTokensUsed)}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              tokens this month (all users)
            </p>
            <div className="mt-3 pt-3 border-t">
              <p className="text-xs text-muted-foreground mb-1">Default limit per user</p>
              {editingLimit ? (
                <div className="flex gap-2 items-center">
                  <Input
                    type="number"
                    min={0}
                    value={limitValue}
                    onChange={(e) => setLimitValue(e.target.value)}
                    placeholder="0 = unlimited"
                    className="h-8 w-32 text-sm"
                    onKeyDown={(e) => e.key === "Enter" && handleSaveLimit()}
                  />
                  <Button size="sm" variant="default" onClick={handleSaveLimit} disabled={saving}>
                    {saving ? "..." : "Save"}
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => setEditingLimit(false)}>
                    Cancel
                  </Button>
                </div>
              ) : (
                <button
                  onClick={() => {
                    setLimitValue(String(stats.DefaultMonthlyTokenLimit));
                    setEditingLimit(true);
                  }}
                  className="text-sm font-medium hover:underline cursor-pointer"
                >
                  {stats.DefaultMonthlyTokenLimit > 0
                    ? formatTokens(stats.DefaultMonthlyTokenLimit)
                    : "Unlimited"}{" "}
                  <span className="text-xs text-muted-foreground">(click to edit)</span>
                </button>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Top senders */}
      {stats.TopGlobalSenders.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Top Global Senders This Week</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {stats.TopGlobalSenders.map((sender, i) => (
                <div key={i} className="flex items-center justify-between">
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
                  <span className="text-sm font-medium tabular-nums">
                    {sender.Count}
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
