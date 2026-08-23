"use client";

import { useState, useEffect } from "react";
import { apiGet } from "@/lib/api";
import type { Digest, ListResponse } from "@/types/api";
import { Badge } from "@/components/ui/badge";
import { Sparkles } from "lucide-react";
import Link from "next/link";

function normalizeTopics(topics: string[]): Set<string> {
  return new Set((topics || []).map((t) => t.toLowerCase().trim()));
}

function getSenderAddresses(digest: Digest): Set<string> {
  const set = new Set<string>();
  for (const item of digest.Items || []) {
    if (item.FromAddress) set.add(item.FromAddress);
  }
  return set;
}

function usePrevDigest(currentId: string) {
  const [prevDigest, setPrevDigest] = useState<Digest | null>(null);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      try {
        const list = await apiGet<ListResponse<Digest>>("digests", { limit: "20" });
        const digests = list.data || [];
        const currentIdx = digests.findIndex((d) => d.ID === currentId);
        if (currentIdx === -1 || currentIdx >= digests.length - 1) return;
        const prev = await apiGet<Digest>(`digests/${digests[currentIdx + 1].ID}`);
        if (!cancelled) setPrevDigest(prev);
      } catch {
        // No previous digest available
      }
    };
    load();
    return () => { cancelled = true; };
  }, [currentId]);

  return prevDigest;
}

export function HighlightedTopics({ currentDigest }: { currentDigest: Digest }) {
  const prevDigest = usePrevDigest(currentDigest.ID);
  const prevTopics = prevDigest ? normalizeTopics(prevDigest.TopTopics) : null;

  return (
    <div className="flex flex-wrap gap-2">
      {(currentDigest.TopTopics || []).map((topic) => {
        const isNew = prevTopics !== null && !prevTopics.has(topic.toLowerCase().trim());
        return (
          <Badge
            key={topic}
            variant="secondary"
            className={isNew ? "ring-1 ring-success/40 bg-success/10 text-success" : ""}
          >
            {isNew && <Sparkles className="size-3 mr-1" />}
            {topic}
          </Badge>
        );
      })}
    </div>
  );
}

export function HighlightedItems({ currentDigest }: { currentDigest: Digest }) {
  const prevDigest = usePrevDigest(currentDigest.ID);
  const prevSenders = prevDigest ? getSenderAddresses(prevDigest) : null;

  return (
    <div className="space-y-2">
      {(currentDigest.Items || []).map((item) => {
        const isNewSender = prevSenders !== null && item.FromAddress && !prevSenders.has(item.FromAddress);
        return (
          <Link
            key={item.ID}
            href={`/inbox/${item.EmailID}`}
            className={`block rounded-lg border p-3 transition-colors hover:bg-accent ${
              isNewSender ? "border-l-2 border-l-green-500" : ""
            }`}
          >
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <p className="text-sm font-medium truncate">
                    {item.Subject || `Email ${item.SortOrder + 1}`}
                  </p>
                  {isNewSender && (
                    <Badge variant="outline" className="shrink-0 text-[10px] text-success border-success/30 px-1.5 py-0">
                      New sender
                    </Badge>
                  )}
                </div>
                {item.FromName && (
                  <p className="text-xs text-muted-foreground truncate">
                    {item.FromName}
                  </p>
                )}
                {item.Summary && (
                  <p className="mt-1 text-xs text-muted-foreground line-clamp-2">
                    {item.Summary}
                  </p>
                )}
              </div>
            </div>
          </Link>
        );
      })}
    </div>
  );
}
