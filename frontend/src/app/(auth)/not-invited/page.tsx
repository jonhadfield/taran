"use client";

import { useState, useEffect } from "react";
import { authClient } from "@/lib/auth-client";
import { apiPost } from "@/lib/api";
import { PublicShell } from "@/components/public-shell";
import { Button } from "@/components/ui/button";
import { APP_NAME } from "@/lib/config";
import { Spinner } from "@/components/ui/spinner";

export default function NotInvitedPage() {
  const [status, setStatus] = useState<"idle" | "loading" | "requested">(
    "idle"
  );
  const [error, setError] = useState<string | null>(null);
  const [waitlistOpen, setWaitlistOpen] = useState<boolean | null>(null);

  useEffect(() => {
    fetch("/api/waitlist-status")
      .then((res) => res.json())
      .then((data: { waitlistEnabled: boolean }) =>
        setWaitlistOpen(data.waitlistEnabled)
      )
      .catch(() => setWaitlistOpen(false));
  }, []);

  const handleRequestAccess = async () => {
    setStatus("loading");
    setError(null);
    try {
      await apiPost("waitlist", {});
      setStatus("requested");
    } catch (e) {
      // 201 and 200 both succeed; a conflict means already on waitlist
      if (e instanceof Error && e.message.includes("409")) {
        setStatus("requested");
      } else if (e instanceof Error && e.message.includes("403")) {
        setWaitlistOpen(false);
        setError(null);
      } else {
        setError("Something went wrong. Please try again.");
        setStatus("idle");
      }
    }
  };

  return (
    <PublicShell>
      <div className="flex w-full flex-col items-center gap-5 text-center">
        {status === "requested" ? (
          <>
            <h2 className="text-2xl font-bold tracking-tight">
              You&apos;re on the waitlist
            </h2>
            <p className="text-muted-foreground">
              We&apos;ll notify you when your access is approved.
            </p>
          </>
        ) : (
          <>
            <h2 className="text-2xl font-bold tracking-tight">
              Invite required
            </h2>
            <p className="text-muted-foreground">
              {APP_NAME} is currently invite-only.
              {waitlistOpen
                ? " Request access below and we'll review your request."
                : " Please check back later."}
            </p>
            {error && <p className="text-sm text-destructive">{error}</p>}
            {waitlistOpen && (
              <Button
                onClick={handleRequestAccess}
                disabled={status === "loading"}
              >
                {status === "loading" && <Spinner className="mr-2" />}
                Request Access
              </Button>
            )}
          </>
        )}

        <Button
          variant="outline"
          className="bg-background"
          onClick={async () => {
            await authClient.signOut();
            // Full document load on purpose: a client-side push would keep the
            // React tree — and any state cached in it from the signed-in
            // session — alive after sign-out.
            // eslint-disable-next-line @next/next/no-location-assign-relative-destination
            window.location.href = "/login";
          }}
        >
          Sign out
        </Button>
      </div>
    </PublicShell>
  );
}
