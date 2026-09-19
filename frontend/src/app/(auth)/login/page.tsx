"use client";

import Image from "next/image";
import { authClient } from "@/lib/auth-client";
import { GitHubIcon } from "@/components/github-icon";
import { PublicShell } from "@/components/public-shell";
import { Button } from "@/components/ui/button";
import { SHOW_MARKETING } from "@/lib/config";

export default function LoginPage() {
  return (
    <PublicShell>
      {SHOW_MARKETING && (
        <p className="max-w-sm text-center text-base text-muted-foreground">
          Open-source email digests. Forward your newsletters, get one brief that
          saves you time.
        </p>
      )}

      <section aria-labelledby="sign-in" className="w-full space-y-3">
        <h2 id="sign-in" className="sr-only">
          Sign in
        </h2>
        <Button
          variant="outline"
          className="w-full bg-background"
          onClick={() =>
            authClient.signIn.social({ provider: "google", callbackURL: "/" })
          }
        >
          <GoogleIcon />
          Continue with Google
        </Button>
        <Button
          variant="outline"
          className="w-full bg-background"
          onClick={() =>
            authClient.signIn.social({ provider: "github", callbackURL: "/" })
          }
        >
          <GitHubIcon className="mr-2 size-4" />
          Continue with GitHub
        </Button>
        <p className="text-center text-sm text-muted-foreground">
          Invite-only — sign in to join the waitlist if you need access
        </p>
      </section>

      {SHOW_MARKETING && (
        <>
          <figure className="w-full overflow-hidden rounded-xl border border-border/60 bg-background/80">
            <Image
              src="/digest-flow.png"
              alt="Forward newsletters, digest with AI, read one brief"
              width={1280}
              height={720}
              className="h-auto w-full"
              priority
            />
          </figure>

          <figure className="digest-sheet w-full space-y-4 p-5 sm:p-6">
            <figcaption className="text-center text-sm text-muted-foreground">
              What a digest looks like
            </figcaption>
            <div className="space-y-1">
              <p className="font-reading text-lg font-medium leading-snug">
                Your Daily Newsletter Digest
              </p>
              <p className="text-xs text-muted-foreground">
                14 February – 15 February · 8 emails
              </p>
            </div>
            <p className="font-reading text-[0.9375rem] leading-relaxed text-foreground/90">
              AI breakthroughs dominated today&apos;s newsletters with major
              announcements from leading labs. Markets reacted positively to
              strong earnings, while new open-source tools gained traction in the
              developer community.
            </p>
            <div>
              <p className="mb-2 text-xs font-medium text-muted-foreground">
                Highlights
              </p>
              <ul className="space-y-2">
                {[
                  "New reasoning model achieves state-of-the-art benchmarks",
                  "Tech earnings beat expectations across the board",
                  "Open-source framework hits 50k GitHub stars",
                ].map((item) => (
                  <li
                    key={item}
                    className="flex items-start gap-2 font-reading text-[0.9375rem] leading-relaxed text-muted-foreground"
                  >
                    <span
                      className="mt-2 size-1 shrink-0 rounded-full bg-primary/50"
                      aria-hidden
                    />
                    {item}
                  </li>
                ))}
              </ul>
            </div>
            <ul className="flex flex-wrap gap-2 pt-1">
              {["AI", "Markets", "Open Source", "Startups"].map((topic) => (
                <li
                  key={topic}
                  className="rounded-md bg-muted px-2 py-0.5 text-xs text-muted-foreground"
                >
                  {topic}
                </li>
              ))}
            </ul>
          </figure>
        </>
      )}
    </PublicShell>
  );
}

function GoogleIcon() {
  return (
    <svg className="mr-2 h-4 w-4" viewBox="0 0 24 24">
      <path
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"
        fill="#4285F4"
      />
      <path
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
        fill="#34A853"
      />
      <path
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
        fill="#FBBC05"
      />
      <path
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
        fill="#EA4335"
      />
    </svg>
  );
}
