import { serverFetch } from "@/lib/server-api";
import { isAdmin } from "@/lib/admin";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { Digest, ListResponse } from "@/types/api";
import { BookOpen } from "lucide-react";
import Link from "next/link";
import { GenerateDigestButton } from "./generate-digest-button";

export default async function DigestsPage() {
  let digests: Digest[] = [];
  const admin = await isAdmin();

  try {
    const res = await serverFetch<ListResponse<Digest>>("digests?limit=50");
    digests = res.data || [];
  } catch {
    // Will show empty state
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-balance">Digests</h1>
        {admin && <GenerateDigestButton />}
      </div>

      {digests.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <BookOpen className="size-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium">No digests yet</h3>
            <p className="text-sm text-muted-foreground mt-1">
              Digests are generated automatically from your incoming emails.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 @sm:grid-cols-2">
          {digests.map((digest) => (
            <Link key={digest.ID} href={`/digests/${digest.ID}`}>
              <Card className="h-full border-l-4 border-l-indigo-500 hover:shadow-md transition-shadow">
                <CardContent className="p-6 space-y-3">
                  <div className="flex items-start justify-between gap-2">
                    <h3 className="font-medium">{digest.Title}</h3>
                    <Badge variant="secondary">{digest.EmailCount} emails</Badge>
                  </div>
                  <p className="text-sm text-muted-foreground line-clamp-3">
                    {digest.Summary}
                  </p>
                  {digest.TopTopics?.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {digest.TopTopics.slice(0, 4).map((topic) => (
                        <Badge key={topic} variant="outline" className="text-xs">
                          {topic}
                        </Badge>
                      ))}
                    </div>
                  )}
                  <p className="text-xs text-muted-foreground">
                    {new Date(digest.PeriodStart).toLocaleDateString()} &ndash;{" "}
                    {new Date(digest.PeriodEnd).toLocaleDateString()}
                  </p>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
