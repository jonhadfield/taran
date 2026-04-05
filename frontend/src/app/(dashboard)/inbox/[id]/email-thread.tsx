"use client";

import { useState } from "react";
import { apiGet } from "@/lib/api";
import type { ThreadEmail } from "@/types/api";
import Link from "next/link";
import { MessageSquare, ChevronDown, ChevronUp } from "lucide-react";
import { formatShortDate } from "@/lib/utils";

interface EmailThreadProps {
  emailId: string;
  threadCount: number;
}

export function EmailThread({ emailId, threadCount }: EmailThreadProps) {
  const [thread, setThread] = useState<ThreadEmail[]>([]);
  const [loading, setLoading] = useState(false);
  const [expanded, setExpanded] = useState(false);

  const handleExpand = () => {
    const next = !expanded;
    setExpanded(next);

    if (next && thread.length === 0) {
      setLoading(true);
      apiGet<ThreadEmail[]>(`emails/${emailId}/thread`)
        .then((data) => setThread(data || []))
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  };

  if (threadCount < 2) return null;

  return (
    <div className="rounded-lg border">
      <button
        onClick={handleExpand}
        className="flex items-center justify-between w-full px-4 py-3 text-sm font-medium hover:bg-accent/50 transition-colors"
      >
        <span className="flex items-center gap-2">
          <MessageSquare className="size-4" />
          {threadCount} messages in conversation
        </span>
        {expanded ? <ChevronUp className="size-4" /> : <ChevronDown className="size-4" />}
      </button>

      {expanded && (
        <div className="border-t divide-y">
          {loading ? (
            <div className="px-4 py-3 text-sm text-muted-foreground">Loading thread...</div>
          ) : (
            thread.map((msg) => {
              const isCurrent = msg.ID === emailId;
              return (
                <div
                  key={msg.ID}
                  className={`px-4 py-3 ${isCurrent ? "bg-accent/30" : ""}`}
                >
                  {isCurrent ? (
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium truncate">
                          {msg.Subject}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {msg.FromName || msg.FromAddress} — current message
                        </p>
                      </div>
                      <span className="text-xs text-muted-foreground whitespace-nowrap shrink-0">
                        {formatShortDate(msg.ReceivedAt)}
                      </span>
                    </div>
                  ) : (
                    <Link
                      href={`/inbox/${msg.ID}`}
                      className="flex items-start justify-between gap-3 hover:underline"
                    >
                      <div className="min-w-0 flex-1">
                        <p className={`text-sm truncate ${msg.IsRead ? "text-muted-foreground" : "font-medium"}`}>
                          {msg.Subject}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {msg.FromName || msg.FromAddress}
                        </p>
                        {msg.Summary && (
                          <p className="text-xs text-muted-foreground mt-0.5 line-clamp-1">
                            {msg.Summary}
                          </p>
                        )}
                      </div>
                      <span className="text-xs text-muted-foreground whitespace-nowrap shrink-0">
                        {formatShortDate(msg.ReceivedAt)}
                      </span>
                    </Link>
                  )}
                </div>
              );
            })
          )}
        </div>
      )}
    </div>
  );
}
