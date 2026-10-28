"use client";

import { useEffect, useState } from "react";
import { authClient } from "@/lib/auth-client";
import { safeCallbackURL } from "@/lib/auth-callback";
import { DemoDigest } from "@/components/demo-digest";
import { GitHubIcon } from "@/components/github-icon";
import { PublicShell } from "@/components/public-shell";
import { Button } from "@/components/ui/button";
import { SHOW_MARKETING } from "@/lib/config";

function callbackURLFromPage(): string {
  if (typeof window === "undefined") return "/";
  return safeCallbackURL(new URLSearchParams(window.location.search).get("callbackURL"));
}

export default function LoginPage() {
  // null until known, so the access note doesn't flash the wrong message.
  const [openRegistration, setOpenRegistration] = useState<boolean | null>(null);

  useEffect(() => {
    fetch("/api/waitlist-status")
      .then((res) => res.json())
      .then((data: { openRegistration?: boolean }) =>
        setOpenRegistration(data.openRegistration === true),
      )
      .catch(() => setOpenRegistration(false));
  }, []);

  const signIn = (provider: "google" | "github") => {
    authClient.signIn.social({ provider, callbackURL: callbackURLFromPage() });
  };

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
          onClick={() => signIn("google")}
        >
          <GoogleIcon />
          Continue with Google
        </Button>
        <Button
          variant="outline"
          className="w-full bg-background"
          onClick={() => signIn("github")}
        >
          <GitHubIcon className="mr-2 size-4" />
          Continue with GitHub
        </Button>
        <p className="min-h-5 text-center text-sm text-muted-foreground">
          {openRegistration === true &&
            "Open for sign-ups: continue with Google or GitHub to create your account"}
          {openRegistration === false &&
            "Invite-only — sign in to join the waitlist if you need access"}
        </p>
      </section>

      {SHOW_MARKETING && <DemoDigest priority />}
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
