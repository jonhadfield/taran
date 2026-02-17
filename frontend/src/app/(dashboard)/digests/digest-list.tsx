"use client";

import { usePolling } from "@/hooks/use-polling";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { Digest, ListResponse } from "@/types/api";
import { BookOpen } from "lucide-react";
import Link from "next/link";

interface DigestListProps {
  initialDigests: Digest[];
}

export function DigestList({ initialDigests }: DigestListProps) {
  const res = usePolling<ListResponse<Digest>>(
    "digests?limit=50",
    { data: initialDigests, total: initialDigests.length }
  );

  const digests = res.data || [];

  return (
    <>
      {digests.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <BookOpen className="size-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium">No digests yet</h3>
            <p className="text-sm text-muted-foreground mt-1">
              Once you start receiving emails, digests will be generated automatically.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {digests.map((digest) => (
            <Link key={digest.ID} href={`/digests/${digest.ID}`}>
              <Card className="border-l-4 border-l-indigo-500 hover:shadow-md transition-shadow">
                <CardContent className="p-6 space-y-3">
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <span>
                      {new Date(digest.PeriodStart).toLocaleDateString()} &ndash;{" "}
                      {new Date(digest.PeriodEnd).toLocaleDateString()}
                    </span>
                    <span>&middot;</span>
                    <span>{digest.EmailCount} emails</span>
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
        </div>
      )}
    </>
  );
}
