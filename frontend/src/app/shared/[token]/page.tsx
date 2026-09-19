import { Badge } from "@/components/ui/badge";
import type { Digest } from "@/types/api";
import { notFound } from "next/navigation";
import Link from "next/link";
import { APP_NAME, SHOW_MARKETING } from "@/lib/config";
import { formatShortDate } from "@/lib/utils";
import { PublicShell } from "@/components/public-shell";
import Image from "next/image";

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
    <PublicShell width="2xl" showBrand={false} className="items-stretch gap-10">
      <header className="flex items-center gap-2.5">
        <Image
          src="/logo.svg"
          alt={APP_NAME}
          width={28}
          height={28}
          className="rounded-md"
        />
        <span className="text-sm font-semibold tracking-tight">{APP_NAME}</span>
      </header>

      <article className="digest-sheet space-y-8 p-5 sm:p-8">
        <header className="space-y-2">
          <h1 className="font-reading text-2xl font-medium leading-snug sm:text-3xl">
            {digest.Title}
          </h1>
          <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
            <span>
              {formatShortDate(digest.PeriodStart)} &ndash;{" "}
              {formatShortDate(digest.PeriodEnd)}
            </span>
            <Badge variant="secondary">{digest.EmailCount} emails</Badge>
          </div>
        </header>

        <section aria-labelledby="summary-heading" className="space-y-3">
          <h2 id="summary-heading" className="text-sm font-medium text-muted-foreground">
            Summary
          </h2>
          <p className="font-reading text-base leading-relaxed">{digest.Summary}</p>
        </section>

        {digest.Highlights?.length > 0 && (
          <section aria-labelledby="highlights-heading" className="space-y-3">
            <h2
              id="highlights-heading"
              className="text-sm font-medium text-muted-foreground"
            >
              Highlights
            </h2>
            <ul className="space-y-3">
              {digest.Highlights.map((highlight, i) => (
                <li
                  key={i}
                  className="flex items-start gap-2.5 font-reading text-base leading-relaxed"
                >
                  <span
                    className="mt-2.5 size-1 shrink-0 rounded-full bg-primary/50"
                    aria-hidden
                  />
                  {highlight}
                </li>
              ))}
            </ul>
          </section>
        )}

        {digest.TopTopics?.length > 0 && (
          <section aria-labelledby="topics-heading" className="space-y-3">
            <h2
              id="topics-heading"
              className="text-sm font-medium text-muted-foreground"
            >
              Top topics
            </h2>
            <ul className="flex flex-wrap gap-2">
              {digest.TopTopics.map((topic) => (
                <li key={topic}>
                  <Badge variant="secondary">{topic}</Badge>
                </li>
              ))}
            </ul>
          </section>
        )}

        {digest.Items?.length > 0 && (
          <section aria-labelledby="emails-heading" className="space-y-3">
            <h2
              id="emails-heading"
              className="text-sm font-medium text-muted-foreground"
            >
              Included emails
            </h2>
            <ul className="divide-y">
              {digest.Items.map((item) => (
                <li key={item.ID} className="py-3 first:pt-0 last:pb-0">
                  <p className="text-sm font-medium">
                    {item.Subject || `Email ${item.SortOrder + 1}`}
                  </p>
                  {item.FromName && (
                    <p className="text-xs text-muted-foreground">{item.FromName}</p>
                  )}
                  {item.Summary && (
                    <p className="mt-1 font-reading text-sm leading-relaxed text-muted-foreground">
                      {item.Summary}
                    </p>
                  )}
                </li>
              ))}
            </ul>
          </section>
        )}
      </article>

      <footer className="border-t pt-6 text-center">
        <p className="text-sm text-muted-foreground">
          Powered by{" "}
          <Link
            href="/login"
            className="font-medium text-foreground hover:underline"
          >
            {APP_NAME}
          </Link>
        </p>
        {SHOW_MARKETING && (
          <p className="mt-1 text-xs text-muted-foreground">
            Open-source newsletter digests that save you time.{" "}
            <Link href="/login" className="underline">
              Sign in
            </Link>
          </p>
        )}
      </footer>
    </PublicShell>
  );
}
