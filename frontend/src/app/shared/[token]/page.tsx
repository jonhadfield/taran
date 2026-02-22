import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Mail } from "lucide-react";
import type { Digest } from "@/types/api";
import { notFound } from "next/navigation";
import Link from "next/link";
import { APP_NAME } from "@/lib/config";

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

export const revalidate = 300;

async function fetchPublicDigest(token: string): Promise<Digest | null> {
  try {
    const res = await fetch(`${BACKEND_URL}/api/public/digests/${token}`);
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

export default async function SharedDigestPage({
  params,
}: {
  params: Promise<{ token: string }>;
}) {
  const { token } = await params;

  const digest = await fetchPublicDigest(token);
  if (!digest) {
    notFound();
  }

  return (
    <div className="min-h-screen bg-muted/40">
      <div className="mx-auto max-w-2xl px-4 py-8">
        <div className="space-y-6">
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
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Included Emails</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {digest.Items.map((item) => (
                    <div key={item.ID} className="border-b last:border-0 pb-3 last:pb-0">
                      <p className="text-sm font-medium">
                        {item.Subject || `Email ${item.SortOrder + 1}`}
                      </p>
                      {item.FromName && (
                        <p className="text-xs text-muted-foreground">
                          {item.FromName}
                        </p>
                      )}
                      {item.Summary && (
                        <p className="mt-1 text-xs text-muted-foreground">
                          {item.Summary}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          <div className="border-t pt-6 text-center">
            <div className="flex items-center justify-center gap-2 text-sm text-muted-foreground">
              <Mail className="h-4 w-4" />
              <span>
                Powered by{" "}
                <Link href="/login" className="font-medium text-foreground hover:underline">
                  {APP_NAME}
                </Link>
              </span>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Get AI-powered digests of your newsletters.{" "}
              <Link href="/login" className="underline">
                Sign up free
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
