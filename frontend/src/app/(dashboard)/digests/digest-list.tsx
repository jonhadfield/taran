"use client";

import { useState } from "react";
import { usePolling } from "@/hooks/use-polling";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { Digest, ListResponse } from "@/types/api";
import { Mail, MailX } from "lucide-react";
import { DigestIllustration } from "@/components/empty-state";
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty";
import Link from "next/link";
import { formatShortDate } from "@/lib/utils";

const PAGE_SIZE = 50;

interface DigestListProps {
  initialDigests: Digest[];
  initialTotal: number;
}

export function DigestList({ initialDigests, initialTotal }: DigestListProps) {
  const [limit, setLimit] = useState(PAGE_SIZE);
  const { data: res } = usePolling<ListResponse<Digest>>(
    `digests?limit=${limit}`,
    { data: initialDigests, total: initialTotal }
  );

  const digests = res.data || [];
  const total = res.total;

  return (
    <>
      {digests.length === 0 ? (
        <Card>
          <CardContent>
            <Empty>
              <EmptyHeader>
                <EmptyMedia><DigestIllustration /></EmptyMedia>
                <EmptyTitle>No digests yet</EmptyTitle>
                <EmptyDescription>Once you start receiving emails, digests will be generated automatically.</EmptyDescription>
              </EmptyHeader>
            </Empty>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {digests.map((digest) => (
            <Link key={digest.ID} href={`/digests/${digest.ID}`}>
              <Card className="border-l-4 border-l-indigo-500 hover:shadow-md transition-shadow">
                <CardContent className="p-4 sm:p-6 space-y-3">
                  <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
                    <span>
                      {formatShortDate(digest.PeriodStart)} &ndash;{" "}
                      {formatShortDate(digest.PeriodEnd)}
                    </span>
                    <span>&middot;</span>
                    <span>{digest.EmailCount} emails</span>
                    <span className="hidden sm:inline">&middot;</span>
                    {digest.SentAt ? (
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
                  <h3 className="font-medium">{digest.Title}</h3>
                  <p className="text-sm text-muted-foreground line-clamp-3">
                    {digest.Summary}
                  </p>
                  {digest.TopTopics?.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {digest.TopTopics.slice(0, 4).map((topic) => (
                        <Badge key={topic} variant="outline" className="text-xs">
                          {topic}
                        </Badge>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </Link>
          ))}
          {digests.length < total && (
            <div className="flex justify-center">
              <Button
                variant="outline"
                onClick={() => setLimit((prev) => prev + PAGE_SIZE)}
              >
                Load more ({total - digests.length} remaining)
              </Button>
            </div>
          )}
        </div>
      )}
    </>
  );
}
