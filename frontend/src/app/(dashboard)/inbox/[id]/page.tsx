import { serverFetch } from "@/lib/server-api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { EmailResponse } from "@/types/api";
import { ArrowLeft, Sparkles, CheckCircle2 } from "lucide-react";
import Link from "next/link";
import { notFound } from "next/navigation";
import { EmailActions } from "./actions";
import { FeedbackButtons } from "./feedback-buttons";
import DOMPurify from "isomorphic-dompurify";

function isSafeURL(url: string): boolean {
  try {
    const parsed = new URL(url, "https://placeholder.invalid");
    return parsed.protocol === "https:" || parsed.protocol === "http:" || parsed.protocol === "mailto:";
  } catch {
    return false;
  }
}

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
      <div className="flex items-center gap-4">
        <Link
          href="/inbox"
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          Back to Inbox
        </Link>
      </div>

      <div className="space-y-4">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold text-balance">{email.Subject}</h1>
            <p className="text-sm text-muted-foreground">
              From: <span className="text-foreground">{email.FromName || email.FromAddress}</span>
              {email.FromName && (
                <span className="ml-1">&lt;{email.FromAddress}&gt;</span>
              )}
            </p>
            <p className="text-sm text-muted-foreground">
              To: {email.ToAddress}
            </p>
            <p className="text-sm text-muted-foreground">
              {new Date(email.ReceivedAt).toLocaleString()}
            </p>
          </div>
          <EmailActions email={email} />
        </div>

        <Separator />

        {email.extraction && (
          <Card className="border-t-2 border-t-indigo-500">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                <Sparkles className="size-5 text-indigo-500" />
                AI Summary
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <p className="text-sm">{email.extraction.Summary}</p>
              </div>

              {email.extraction.KeyPoints?.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium mb-2">Key Points</h4>
                  <ul className="space-y-1.5">
                    {email.extraction.KeyPoints.map((point, i) => (
                      <li key={i} className="flex items-start gap-2 text-sm text-muted-foreground">
                        <CheckCircle2 className="size-4 mt-0.5 shrink-0 text-indigo-500" />
                        <span>{point}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {email.extraction.Topics?.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium mb-2">Topics</h4>
                  <div className="flex flex-wrap gap-2">
                    {email.extraction.Topics.map((topic) => (
                      <Badge key={topic} variant="secondary">
                        {topic}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}

              {email.extraction.ActionItems?.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium mb-2">Action Items</h4>
                  <ul className="space-y-1.5">
                    {email.extraction.ActionItems.map((item, i) => (
                      <li key={i} className="flex items-start gap-2 text-sm text-muted-foreground">
                        <CheckCircle2 className="size-4 mt-0.5 shrink-0 text-amber-500" />
                        <span>{item}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {email.extraction.Links?.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium mb-2">Links</h4>
                  <ul className="space-y-1 text-sm">
                    {email.extraction.Links.map((link, i) => (
                      <li key={i}>
                        <a
                          href={isSafeURL(link.url) ? link.url : "#"}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          {link.title || link.url}
                        </a>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {email.extraction.Sentiment && (
                <div>
                  <h4 className="text-sm font-medium mb-1">Sentiment</h4>
                  <Badge variant="outline">{email.extraction.Sentiment}</Badge>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {email.extraction && (
          <FeedbackButtons emailId={id} />
        )}

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Email Content</CardTitle>
          </CardHeader>
          <CardContent>
            {email.HTMLBody ? (
              <div
                className="prose prose-sm max-w-none dark:prose-invert"
                dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(email.HTMLBody, { FORBID_TAGS: ["style"] }) }}
              />
            ) : (
              <pre className="whitespace-pre-wrap text-sm font-mono">
                {email.TextBody}
              </pre>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
