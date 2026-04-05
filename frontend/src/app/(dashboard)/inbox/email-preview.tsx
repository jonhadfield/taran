"use client";

import { useState, useEffect, useRef } from "react";
import { apiGet, apiPatch } from "@/lib/api";
import type { EmailResponse } from "@/types/api";
import { Separator } from "@/components/ui/separator";
import { Sparkles, Clock, ExternalLink, Inbox, Loader2 } from "lucide-react";
import { ExtractionSummary } from "@/components/extraction-summary";
import Link from "next/link";
import { EmailActions } from "./[id]/actions";
import { EmailLabels } from "./[id]/email-labels";
import { FeedbackButtons } from "./[id]/feedback-buttons";
import { formatDateTime, estimateReadingTime } from "@/lib/utils";

interface EmailPreviewProps {
  id: string | null;
  onDeleted?: () => void;
}

export function EmailPreview({ id, onDeleted }: EmailPreviewProps) {
  const [state, setState] = useState<{
    email: EmailResponse | null;
    loading: boolean;
    error: boolean;
    fetchedId: string | null;
  }>({ email: null, loading: false, error: false, fetchedId: null });
  const markedRead = useRef<string | null>(null);

  // Fetch when id changes and doesn't match the last fetched id
  if (id !== state.fetchedId) {
    if (!id) {
      setState({ email: null, loading: false, error: false, fetchedId: null });
    } else {
      setState({ email: null, loading: true, error: false, fetchedId: id });
    }
  }

  useEffect(() => {
    if (!id) return;

    let cancelled = false;

    apiGet<EmailResponse>(`emails/${id}`)
      .then((data) => {
        if (cancelled) return;
        setState({ email: data, loading: false, error: false, fetchedId: id });

        // Mark as read (once per email)
        if (!data.IsRead && markedRead.current !== id) {
          markedRead.current = id;
          apiPatch(`emails/${id}`, { IsRead: true }).catch(() => {});
        }
      })
      .catch(() => {
        if (!cancelled) setState({ email: null, loading: false, error: true, fetchedId: id });
      });

    return () => { cancelled = true; };
  }, [id]);

  const { email, loading, error } = state;

  if (!id) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center p-8">
        <Inbox className="size-12 text-muted-foreground mb-4" />
        <h3 className="text-lg font-medium">Select an email</h3>
        <p className="text-sm text-muted-foreground mt-1">
          Choose an email from the list to preview it here.
        </p>
        <p className="text-xs text-muted-foreground mt-3">
          <kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">j</kbd>/<kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">k</kbd> to navigate
          {" "}<kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">Enter</kbd> to preview
        </p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error || !email) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center p-8">
        <p className="text-sm text-muted-foreground">Failed to load email.</p>
      </div>
    );
  }

  const ext = email.extraction;

  return (
    <div className="p-4 space-y-4">
      {/* Header with open link */}
      <div className="flex items-start justify-between gap-2">
        <div className="space-y-1 min-w-0">
          <h2 className="text-lg font-bold text-balance break-words">{email.Subject}</h2>
          <p className="text-sm text-muted-foreground">
            From: <span className="text-foreground">{email.FromName || email.FromAddress}</span>
            {email.FromName && (
              <span className="ml-1">&lt;{email.FromAddress}&gt;</span>
            )}
          </p>
          <p className="text-sm text-muted-foreground flex items-center gap-2 flex-wrap">
            <span>{formatDateTime(email.ReceivedAt)}</span>
            {email.TextBody && (
              <span className="inline-flex items-center gap-1 text-xs">
                <Clock className="size-3" />
                {estimateReadingTime(email.TextBody)}
              </span>
            )}
          </p>
        </div>
        <Link
          href={`/inbox/${email.ID}`}
          className="inline-flex items-center gap-1 shrink-0 rounded-md border px-2.5 py-1.5 text-xs font-medium hover:bg-accent transition-colors"
        >
          <ExternalLink className="size-3" />
          Full view
        </Link>
      </div>

      {/* Actions */}
      <EmailActions email={email} onDeleted={onDeleted} />

      {/* Labels */}
      <EmailLabels emailId={email.ID} />

      <Separator />

      {/* AI Summary */}
      {ext && (
        <div className="space-y-3">
          <h3 className="flex items-center gap-2 text-sm font-semibold">
            <Sparkles className="size-4 text-indigo-500" />
            AI Summary
          </h3>
          <ExtractionSummary extraction={ext} compact />
        </div>
      )}

      {ext && (
        <>
          <Separator />
          <FeedbackButtons emailId={email.ID} />
        </>
      )}

      {/* Text body preview — only when there's no AI extraction */}
      {!ext && email.TextBody && (
        <>
          <Separator />
          <div>
            <h4 className="text-xs font-medium text-muted-foreground mb-2">Content Preview</h4>
            <p className="text-sm text-muted-foreground whitespace-pre-wrap line-clamp-[12]">
              {email.TextBody.slice(0, 800)}
            </p>
          </div>
        </>
      )}
    </div>
  );
}
