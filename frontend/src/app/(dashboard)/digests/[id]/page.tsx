import { serverFetch } from "@/lib/server-api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { Digest, UserPreference } from "@/types/api";
import { ArrowLeft, Mail, MailX, Clock } from "lucide-react";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ShareButton } from "./share-button";
import { SendEmailButton } from "./send-email-button";
import { DeleteDigestButton } from "./delete-button";
import { DigestFeedbackButtons } from "./digest-feedback-buttons";
import { DigestComparison } from "./digest-comparison";
import { isAdmin } from "@/lib/admin";

export default async function DigestDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let digest: Digest;
  try {
    digest = await serverFetch<Digest>(`digests/${id}`);
  } catch {
    notFound();
  }

  const [admin, prefs] = await Promise.all([
    isAdmin(),
    serverFetch<UserPreference>("preferences").catch(() => null),
  ]);

  return (
    <div className="space-y-6">
      <Link
        href="/digests"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Digests
      </Link>

      <div className="space-y-2">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
          <h1 className="text-2xl font-bold">{digest.Title}</h1>
          <div className="flex items-center gap-2">
            {admin && (
              <SendEmailButton digestId={digest.ID} alreadySent={!!digest.SentAt} />
            )}
            <ShareButton digestId={digest.ID} initialToken={digest.ShareToken} />
            <DeleteDigestButton digestId={digest.ID} />
          </div>
        </div>
        <DigestFeedbackButtons digestId={digest.ID} />
        <div className="flex items-center gap-3 text-sm text-muted-foreground">
          <span>
            {new Date(digest.PeriodStart).toLocaleDateString()} &ndash;{" "}
            {new Date(digest.PeriodEnd).toLocaleDateString()}
          </span>
          <Badge variant="secondary">{digest.EmailCount} emails</Badge>
        </div>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          {digest.SentAt ? (
            <>
              <Mail className="size-4 text-green-600" />
              <span>Emailed on {new Date(digest.SentAt).toLocaleString()}</span>
            </>
          ) : prefs?.DigestEmail ? (
            <>
              <Clock className="size-4" />
              <span>
                Scheduled for{" "}
                {new Date(
                  2000, 0, 1, prefs.DigestHour
                ).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })}{" "}
                {prefs.DigestTimezone}
              </span>
            </>
          ) : (
            <>
              <MailX className="size-4" />
              <span>
                Email delivery not enabled &mdash;{" "}
                <Link href="/settings" className="underline hover:text-foreground">
                  enable in settings
                </Link>
              </span>
            </>
          )}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm">{digest.Summary}</p>
        </CardContent>
      </Card>

      {digest.Highlights?.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Highlights</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="list-disc list-inside space-y-2 text-sm">
              {digest.Highlights.map((highlight, i) => (
                <li key={i}>{highlight}</li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}

      {digest.TopTopics?.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Top Topics</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              {digest.TopTopics.map((topic) => (
                <Badge key={topic} variant="secondary">
                  {topic}
                </Badge>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      <DigestComparison currentDigest={digest} />

      {digest.Items?.length > 0 && (
        <>
          <Separator />
          <div className="space-y-2">
            <h2 className="text-lg font-medium">Included Emails</h2>
            <div className="space-y-2">
              {digest.Items.map((item) => (
                <Link
                  key={item.ID}
                  href={`/inbox/${item.EmailID}`}
                  className="block rounded-lg border p-3 transition-colors hover:bg-accent"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium truncate">
                        {item.Subject || `Email ${item.SortOrder + 1}`}
                      </p>
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
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
