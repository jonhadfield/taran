import { serverFetch } from "@/lib/server-api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { Digest } from "@/types/api";
import { ArrowLeft } from "lucide-react";
import Link from "next/link";
import { notFound } from "next/navigation";

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
        <h1 className="text-2xl font-bold">{digest.Title}</h1>
        <div className="flex items-center gap-3 text-sm text-muted-foreground">
          <span>
            {new Date(digest.PeriodStart).toLocaleDateString()} &ndash;{" "}
            {new Date(digest.PeriodEnd).toLocaleDateString()}
          </span>
          <Badge variant="secondary">{digest.EmailCount} emails</Badge>
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
                  className="block rounded-lg border p-3 text-sm transition-colors hover:bg-accent"
                >
                  View email {item.SortOrder + 1}
                </Link>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
