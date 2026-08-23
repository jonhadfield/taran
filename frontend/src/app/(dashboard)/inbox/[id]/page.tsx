import { serverFetch } from "@/lib/server-api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { EmailResponse } from "@/types/api";
import { Sparkles, Info, Paperclip, Clock } from "lucide-react";
import { ExtractionSummary } from "@/components/extraction-summary";
import Link from "next/link";
import { notFound } from "next/navigation";
import { EmailActions } from "./actions";
import { EmailLabels } from "./email-labels";
import { FeedbackButtons } from "./feedback-buttons";
import { ReprocessButton } from "./reprocess-button";
import { UnsubscribeButton } from "./unsubscribe-button";
import { EmailThread } from "./email-thread";
import { EmailContentCard } from "./email-content-card";
import { formatDateTime, estimateReadingTime, pluralize } from "@/lib/utils";

export default async function EmailDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let email: EmailResponse;
  try {
    email = await serverFetch<EmailResponse>(`emails/${id}`);
  } catch {
    notFound();
  }

  return (
    <div className="space-y-6">
      <nav className="flex items-center gap-1.5 text-sm text-muted-foreground">
        <Link href="/inbox" className="hover:text-foreground transition-colors">Inbox</Link>
        <span>/</span>
        <span className="text-foreground truncate max-w-xs">{email.Subject}</span>
      </nav>

      <div className="space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3 sm:gap-4">
          <div className="space-y-1 min-w-0">
            <h1 className="text-xl sm:text-2xl font-bold text-balance break-words">{email.Subject}</h1>
            <p className="text-sm text-muted-foreground">
              From: <span className="text-foreground">{email.FromName || email.FromAddress}</span>
              {email.FromName && (
                <span className="ml-1">&lt;{email.FromAddress}&gt;</span>
              )}
            </p>
            <p className="text-sm text-muted-foreground">
              To: {email.ToAddress}
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
          <EmailActions email={email} />
        </div>

        <EmailLabels emailId={id} />

        {(email.UnsubscribeURL || email.UnsubscribeMailto) && (
          <UnsubscribeButton emailId={id} />
        )}

        {email.ThreadCount && email.ThreadCount > 1 && (
          <EmailThread emailId={id} threadCount={email.ThreadCount} />
        )}

        <Separator />

        {!email.Extraction && (email.Status === "skipped" || email.Status === "failed") && (
          <div className="flex items-start gap-3 rounded-lg border border-muted bg-muted/50 p-4">
            <Info className="size-5 shrink-0 text-muted-foreground mt-0.5" />
            <div>
              <p className="text-sm font-medium">
                {email.Status === "skipped" ? "Email skipped" : "Processing failed"}
              </p>
              {email.StatusReason && (
                <p className="text-sm text-muted-foreground">{email.StatusReason}</p>
              )}
              {email.Status === "failed" && email.RetryCount > 0 && email.RetryCount < 5 && (
                <p className="text-xs text-muted-foreground mt-1">
                  Retried {email.RetryCount} {pluralize(email.RetryCount, "time")}, will retry again automatically
                </p>
              )}
              {email.Status === "failed" && email.RetryCount >= 5 && (
                <p className="text-xs text-muted-foreground mt-1">
                  Auto-retry exhausted after {email.RetryCount} attempts
                </p>
              )}
              <ReprocessButton emailId={id} />
            </div>
          </div>
        )}

        {email.Extraction && (
          <Card className="border-t-2 border-t-indigo-500">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                <Sparkles className="size-5 text-info" />
                AI Summary
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ExtractionSummary extraction={email.Extraction} />
            </CardContent>
          </Card>
        )}

        {email.Extraction && (
          <FeedbackButtons emailId={id} />
        )}

        {email.Attachments && email.Attachments.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                <Paperclip className="size-5" />
                Attachments ({email.Attachments.length})
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {email.Attachments.map((att) => (
                  <div
                    key={att.ID}
                    className="flex items-center justify-between gap-2 rounded-lg border px-3 py-2"
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      <Paperclip className="size-4 shrink-0 text-muted-foreground" />
                      <span className="text-sm truncate">{att.Filename}</span>
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <Badge variant="outline" className="text-xs hidden sm:inline-flex">
                        {att.ContentType.split("/").pop()}
                      </Badge>
                      <span className="text-xs text-muted-foreground">
                        {att.SizeBytes < 1024
                          ? `${att.SizeBytes} B`
                          : att.SizeBytes < 1024 * 1024
                            ? `${(att.SizeBytes / 1024).toFixed(1)} KB`
                            : `${(att.SizeBytes / (1024 * 1024)).toFixed(1)} MB`}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        <EmailContentCard
          htmlBody={email.HTMLBody}
          textBody={email.TextBody}
        />
      </div>
    </div>
  );
}
