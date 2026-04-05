import { serverFetch } from "@/lib/server-api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { WeeklySummary } from "@/types/api";
import { Mail, MailX, CheckSquare } from "lucide-react";
import Link from "next/link";
import { notFound } from "next/navigation";
import { formatShortDate, formatDateTime } from "@/lib/utils";
import { CATEGORY_COLORS } from "@/lib/category-constants";

export default async function WeeklySummaryDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let summary: WeeklySummary;
  try {
    summary = await serverFetch<WeeklySummary>(`weekly-summaries/${id}`);
  } catch {
    notFound();
  }

  const totalCategoryEmails = summary.Categories
    ? Object.values(summary.Categories).reduce((a, b) => a + b, 0)
    : 0;

  return (
    <div className="space-y-6">
      <nav className="flex items-center gap-1.5 text-sm text-muted-foreground">
        <Link
          href="/digests"
          className="hover:text-foreground transition-colors"
        >
          Digests
        </Link>
        <span>/</span>
        <span className="text-foreground">Weekly Summary</span>
      </nav>

      <div className="space-y-2">
        <h1 className="text-2xl font-bold">Weekly Summary</h1>
        <div className="flex items-center gap-3 text-sm text-muted-foreground">
          <span>
            {formatShortDate(summary.PeriodStart)} &ndash;{" "}
            {formatShortDate(summary.PeriodEnd)}
          </span>
          <Badge variant="secondary">{summary.EmailCount} emails</Badge>
        </div>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          {summary.SentAt ? (
            <>
              <Mail className="size-4 text-green-600" />
              <span>Emailed on {formatDateTime(summary.SentAt)}</span>
            </>
          ) : (
            <>
              <MailX className="size-4" />
              <span>Not emailed</span>
            </>
          )}
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Overview</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Emails received</span>
              <span className="font-medium">{summary.EmailCount}</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Action items</span>
              <span className="font-medium flex items-center gap-1">
                <CheckSquare className="size-3.5" />
                {summary.ActionItems}
              </span>
            </div>
          </CardContent>
        </Card>

        {summary.Categories &&
          Object.keys(summary.Categories).length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Categories</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {Object.entries(summary.Categories)
                  .sort(([, a], [, b]) => b - a)
                  .map(([category, count]) => (
                    <div key={category} className="space-y-1">
                      <div className="flex items-center justify-between text-sm">
                        <Badge
                          className={`text-xs ${CATEGORY_COLORS[category] || CATEGORY_COLORS.other}`}
                        >
                          {category}
                        </Badge>
                        <span className="text-muted-foreground">
                          {count} email{count !== 1 ? "s" : ""}
                        </span>
                      </div>
                      {totalCategoryEmails > 0 && (
                        <div className="h-1.5 rounded-full bg-muted overflow-hidden">
                          <div
                            className="h-full rounded-full bg-emerald-500"
                            style={{
                              width: `${(count / totalCategoryEmails) * 100}%`,
                            }}
                          />
                        </div>
                      )}
                    </div>
                  ))}
              </CardContent>
            </Card>
          )}
      </div>

      {summary.TopSenders?.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Top Senders</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="divide-y">
              {summary.TopSenders.map((sender) => (
                <div
                  key={sender.FromAddress}
                  className="flex items-center justify-between py-2 text-sm"
                >
                  <div className="min-w-0">
                    <p className="font-medium truncate">
                      {sender.FromName || sender.FromAddress}
                    </p>
                    {sender.FromName && (
                      <p className="text-xs text-muted-foreground truncate">
                        {sender.FromAddress}
                      </p>
                    )}
                  </div>
                  <Badge variant="secondary" className="ml-2 shrink-0">
                    {sender.Count} email{sender.Count !== 1 ? "s" : ""}
                  </Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
