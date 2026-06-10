"use client";

import { useState } from "react";
import { usePolling } from "@/hooks/use-polling";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { WeeklySummary, ListResponse } from "@/types/api";
import { Mail, MailX } from "lucide-react";
import { DigestIllustration } from "@/components/empty-state";
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty";
import Link from "next/link";
import { formatShortDate } from "@/lib/utils";
import { CATEGORY_COLORS } from "@/lib/category-constants";

const PAGE_SIZE = 50;

interface WeeklySummaryListProps {
  initialSummaries: WeeklySummary[];
  initialTotal: number;
}

export function WeeklySummaryList({
  initialSummaries,
  initialTotal,
}: WeeklySummaryListProps) {
  const [limit, setLimit] = useState(PAGE_SIZE);
  const { data: res } = usePolling<ListResponse<WeeklySummary>>(
    `weekly-summaries?limit=${limit}`,
    { data: initialSummaries, total: initialTotal }
  );

  const summaries = res.data || [];
  const total = res.total;

  return (
    <>
      {summaries.length === 0 ? (
        <Card>
          <CardContent>
            <Empty>
              <EmptyHeader>
                <EmptyMedia><DigestIllustration /></EmptyMedia>
                <EmptyTitle>No weekly summaries yet</EmptyTitle>
                <EmptyDescription>Weekly summaries are generated automatically at the end of each week.</EmptyDescription>
              </EmptyHeader>
            </Empty>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {summaries.map((summary) => (
            <Link key={summary.ID} href={`/digests/weekly/${summary.ID}`}>
              <Card className="border-l-4 border-l-emerald-500 hover:shadow-md transition-shadow">
                <CardContent className="p-4 sm:p-6 space-y-3">
                  <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
                    <span>
                      {formatShortDate(summary.PeriodStart)} &ndash;{" "}
                      {formatShortDate(summary.PeriodEnd)}
                    </span>
                    <span>&middot;</span>
                    <span>{summary.EmailCount} emails</span>
                    <span className="hidden sm:inline">&middot;</span>
                    {summary.SentAt ? (
                      <span className="hidden sm:inline-flex items-center gap-1">
                        <Mail className="size-3" />
                        Emailed
                      </span>
                    ) : (
                      <span className="hidden sm:inline-flex items-center gap-1 text-muted-foreground">
                        <MailX className="size-3" />
                        Not emailed
                      </span>
                    )}
                  </div>

                  <div className="flex items-center gap-2">
                    <h3 className="font-medium">
                      Weekly Summary
                    </h3>
                  </div>

                  {summary.TopSenders?.length > 0 && (
                    <div className="flex flex-wrap items-center gap-1 text-sm text-muted-foreground">
                      <span className="text-xs">Top senders:</span>
                      {summary.TopSenders.slice(0, 3).map((sender) => (
                        <Badge
                          key={sender.FromAddress}
                          variant="outline"
                          className="text-xs"
                        >
                          {sender.FromName || sender.FromAddress}
                        </Badge>
                      ))}
                    </div>
                  )}

                  {summary.Categories &&
                    Object.keys(summary.Categories).length > 0 && (
                      <div className="flex flex-wrap gap-1">
                        {Object.entries(summary.Categories).map(
                          ([category, count]) => (
                            <Badge
                              key={category}
                              className={`text-xs ${CATEGORY_COLORS[category] || CATEGORY_COLORS.other}`}
                            >
                              {category} ({count})
                            </Badge>
                          )
                        )}
                      </div>
                    )}
                </CardContent>
              </Card>
            </Link>
          ))}
          {summaries.length < total && (
            <div className="flex justify-center">
              <Button
                variant="outline"
                onClick={() => setLimit((prev) => prev + PAGE_SIZE)}
              >
                Load more ({total - summaries.length} remaining)
              </Button>
            </div>
          )}
        </div>
      )}
    </>
  );
}
