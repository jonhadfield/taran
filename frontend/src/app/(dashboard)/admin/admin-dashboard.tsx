"use client";

import { useState } from "react";
import { toast } from "sonner";
import { usePolling } from "@/hooks/use-polling";
import { apiPatch } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import type { AdminStats } from "@/types/api";
import { Users, Mail, BookOpen, TrendingUp, Cpu, Gauge } from "lucide-react";

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(0)}K`;
  return String(n);
}

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
