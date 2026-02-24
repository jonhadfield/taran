"use client";

import { usePolling } from "@/hooks/use-polling";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { AdminStats } from "@/types/api";
import { Users, Mail, BookOpen, TrendingUp, Cpu } from "lucide-react";

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
};

export function AdminDashboard() {
  const { data: stats } = usePolling<AdminStats>("admin/stats", emptyStats, 30_000);

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
