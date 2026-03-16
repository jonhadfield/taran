"use client";

import { useState, useEffect } from "react";
import { apiGet } from "@/lib/api";
import type { Digest, ListResponse } from "@/types/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ArrowUp, ArrowDown, Plus, Minus, TrendingUp, Users, Hash } from "lucide-react";

interface DigestComparisonProps {
  currentDigest: Digest;
}

interface SenderInfo {
  fromAddress: string;
  fromName: string;
  count: number;
}

function getSenders(digest: Digest): Map<string, SenderInfo> {
  const map = new Map<string, SenderInfo>();
  for (const item of digest.Items || []) {
    const addr = item.FromAddress || "";
    if (!addr) continue;
    const existing = map.get(addr);
    if (existing) {
      existing.count++;
    } else {
      map.set(addr, { fromAddress: addr, fromName: item.FromName || addr, count: 1 });
    }
  }
  return map;
}

function normalizeTopics(topics: string[]): Set<string> {
  return new Set((topics || []).map((t) => t.toLowerCase().trim()));
}

export function DigestComparison({ currentDigest }: DigestComparisonProps) {
  const [prevDigest, setPrevDigest] = useState<Digest | null>(null);
  const [loading, setLoading] = useState(true);
  const [noPrevious, setNoPrevious] = useState(false);

  useEffect(() => {
    const load = async () => {
      try {
        // Fetch recent digests to find the previous one
        const list = await apiGet<ListResponse<Digest>>("digests", { limit: "20" });
        const digests = list.data || [];

        // Find the index of the current digest, then get the next one (previous chronologically)
        const currentIdx = digests.findIndex((d) => d.ID === currentDigest.ID);
        if (currentIdx === -1 || currentIdx >= digests.length - 1) {
          setNoPrevious(true);
          return;
        }

        const prevId = digests[currentIdx + 1].ID;
        const prev = await apiGet<Digest>(`digests/${prevId}`);
        setPrevDigest(prev);
      } catch {
        setNoPrevious(true);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [currentDigest.ID]);

  if (loading || noPrevious || !prevDigest) return null;

  const currentSenders = getSenders(currentDigest);
  const prevSenders = getSenders(prevDigest);
  const currentTopics = normalizeTopics(currentDigest.TopTopics);
  const prevTopics = normalizeTopics(prevDigest.TopTopics);

  // New and dropped senders
  const newSenders: SenderInfo[] = [];
  for (const [addr, info] of currentSenders) {
    if (!prevSenders.has(addr)) newSenders.push(info);
  }
  const droppedSenders: SenderInfo[] = [];
  for (const [addr, info] of prevSenders) {
    if (!currentSenders.has(addr)) droppedSenders.push(info);
  }

  // New and dropped topics (compare using original casing for display)
  const newTopics = (currentDigest.TopTopics || []).filter(
    (t) => !prevTopics.has(t.toLowerCase().trim())
  );
  const droppedTopics = (prevDigest.TopTopics || []).filter(
    (t) => !currentTopics.has(t.toLowerCase().trim())
  );

  const emailDelta = currentDigest.EmailCount - prevDigest.EmailCount;

  const hasChanges = newSenders.length > 0 || droppedSenders.length > 0 || newTopics.length > 0 || droppedTopics.length > 0;
  if (!hasChanges && emailDelta === 0) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg flex items-center gap-2">
          <TrendingUp className="size-5" />
          Changes from Previous Digest
        </CardTitle>
        <p className="text-xs text-muted-foreground">
          Compared to {new Date(prevDigest.PeriodStart).toLocaleDateString()} &ndash;{" "}
          {new Date(prevDigest.PeriodEnd).toLocaleDateString()}
        </p>
      </CardHeader>
      <CardContent className="space-y-5">
        {/* Email count change */}
        {emailDelta !== 0 && (
          <div className="flex items-center gap-2 text-sm">
            {emailDelta > 0 ? (
              <ArrowUp className="size-4 text-green-600" />
            ) : (
              <ArrowDown className="size-4 text-orange-600" />
            )}
            <span>
              <span className="font-medium">
                {Math.abs(emailDelta)} {emailDelta > 0 ? "more" : "fewer"}
              </span>{" "}
              email{Math.abs(emailDelta) !== 1 ? "s" : ""} than last digest
              ({prevDigest.EmailCount} → {currentDigest.EmailCount})
            </span>
          </div>
        )}

        {/* Topics */}
        {(newTopics.length > 0 || droppedTopics.length > 0) && (
          <div className="space-y-2">
            <h3 className="text-sm font-medium flex items-center gap-1.5">
              <Hash className="size-3.5" />
              Topics
            </h3>
            {newTopics.length > 0 && (
              <div className="flex flex-wrap items-center gap-1.5">
                <span className="text-xs text-green-600 font-medium flex items-center gap-0.5">
                  <Plus className="size-3" />
                  New
                </span>
                {newTopics.map((topic) => (
                  <Badge key={topic} variant="secondary" className="text-green-700 dark:text-green-400 bg-green-50 dark:bg-green-950/30">
                    {topic}
                  </Badge>
                ))}
              </div>
            )}
            {droppedTopics.length > 0 && (
              <div className="flex flex-wrap items-center gap-1.5">
                <span className="text-xs text-muted-foreground font-medium flex items-center gap-0.5">
                  <Minus className="size-3" />
                  Gone
                </span>
                {droppedTopics.map((topic) => (
                  <Badge key={topic} variant="secondary" className="text-muted-foreground line-through">
                    {topic}
                  </Badge>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Senders */}
        {(newSenders.length > 0 || droppedSenders.length > 0) && (
          <div className="space-y-2">
            <h3 className="text-sm font-medium flex items-center gap-1.5">
              <Users className="size-3.5" />
              Senders
            </h3>
            {newSenders.length > 0 && (
              <div className="space-y-1">
                <span className="text-xs text-green-600 font-medium flex items-center gap-0.5">
                  <Plus className="size-3" />
                  New in this digest
                </span>
                <div className="flex flex-wrap gap-1.5">
                  {newSenders.map((s) => (
                    <Badge key={s.fromAddress} variant="outline" className="text-green-700 dark:text-green-400 border-green-200 dark:border-green-800">
                      {s.fromName}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
            {droppedSenders.length > 0 && (
              <div className="space-y-1">
                <span className="text-xs text-muted-foreground font-medium flex items-center gap-0.5">
                  <Minus className="size-3" />
                  Not in this digest
                </span>
                <div className="flex flex-wrap gap-1.5">
                  {droppedSenders.map((s) => (
                    <Badge key={s.fromAddress} variant="outline" className="text-muted-foreground">
                      {s.fromName}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
