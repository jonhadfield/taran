import { serverFetch } from "@/lib/server-api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { Digest, UserPreference } from "@/types/api";
import { Mail, MailX, Clock } from "lucide-react";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ShareButton } from "./share-button";
import { SendEmailButton } from "./send-email-button";
import { DeleteDigestButton } from "./delete-button";
import { DigestFeedbackButtons } from "./digest-feedback-buttons";
import { DigestComparison } from "./digest-comparison";
import { HighlightedTopics, HighlightedItems } from "./digest-highlights";
import { DownloadPdfButton } from "./download-pdf-button";
import { isAdmin } from "@/lib/admin";
import { formatShortDate, formatDateTime } from "@/lib/utils";
import {
  normalizeDigestDensity,
  normalizeDigestSectionOrder,
} from "@/lib/digest-template";

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
  const sectionOrder = normalizeDigestSectionOrder(prefs?.DigestSectionOrder);
  const density = normalizeDigestDensity(prefs?.DigestDensity);
  const stackClass = density === "compact" ? "space-y-3" : "space-y-6";

  return (
    <div className={stackClass}>
      <nav className="flex items-center gap-1.5 text-sm text-muted-foreground">
        <Link href="/digests" className="hover:text-foreground transition-colors">Digests</Link>
        <span>/</span>
        <span className="text-foreground truncate max-w-xs">{digest.Title}</span>
      </nav>

      <div className="space-y-2">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
          <h1 className="text-2xl font-bold">{digest.Title}</h1>
          <div className="flex items-center gap-2">
            {admin && (
              <SendEmailButton digestId={digest.ID} alreadySent={!!digest.SentAt} />
            )}
            <ShareButton digestId={digest.ID} initialToken={digest.ShareToken} />
            <DownloadPdfButton
              digest={digest}
              sectionOrder={sectionOrder}
              density={density}
            />
            <DeleteDigestButton digestId={digest.ID} />
          </div>
        </div>
        <DigestFeedbackButtons digestId={digest.ID} />
        <div className="flex items-center gap-3 text-sm text-muted-foreground">
          <span>
            {formatShortDate(digest.PeriodStart)} &ndash;{" "}
            {formatShortDate(digest.PeriodEnd)}
          </span>
          <Badge variant="secondary">{digest.EmailCount} emails</Badge>
        </div>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          {digest.SentAt ? (
            <>
              <Mail className="size-4 text-success" />
              <span>Emailed on {formatDateTime(digest.SentAt)}</span>
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
                <Link href="/settings/digest#delivery" className="underline hover:text-foreground">
                  enable in settings
                </Link>
              </span>
            </>
          )}
        </div>
      </div>

      <Card>
        <CardHeader className={density === "compact" ? "pb-2" : undefined}>
          <CardTitle className="text-lg">Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="font-reading text-base leading-relaxed">{digest.Summary}</p>
        </CardContent>
      </Card>

      {sectionOrder.map((section) => {
        if (section === "highlights" && digest.Highlights?.length > 0) {
          return (
            <Card key="highlights">
              <CardHeader className={density === "compact" ? "pb-2" : undefined}>
                <CardTitle className="text-lg">Highlights</CardTitle>
              </CardHeader>
              <CardContent>
                <ul
                  className={
                    density === "compact"
                      ? "list-disc list-inside space-y-1 text-sm"
                      : "list-disc list-inside space-y-2 text-sm"
                  }
                >
                  {digest.Highlights.map((highlight, i) => (
                    <li key={i}>{highlight}</li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          );
        }
        if (section === "topics" && digest.TopTopics?.length > 0) {
          return (
            <Card key="topics">
              <CardHeader className={density === "compact" ? "pb-2" : undefined}>
                <CardTitle className="text-lg">Top Topics</CardTitle>
              </CardHeader>
              <CardContent>
                <HighlightedTopics currentDigest={digest} />
              </CardContent>
            </Card>
          );
        }
        if (section === "emails" && digest.Items?.length > 0) {
          return (
            <div key="emails" className="space-y-2">
              <Separator />
              <h2 className="text-lg font-medium">Included Emails</h2>
              <HighlightedItems currentDigest={digest} />
            </div>
          );
        }
        return null;
      })}

      <DigestComparison currentDigest={digest} />
    </div>
  );
}
