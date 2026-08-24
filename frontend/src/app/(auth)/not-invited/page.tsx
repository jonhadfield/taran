"use client";

import { useState, useEffect } from "react";
import { authClient } from "@/lib/auth-client";
import { apiPost } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { APP_NAME } from "@/lib/config";
import { ShieldX, CheckCircle2 } from "lucide-react";
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
    <div className="flex min-h-screen flex-col items-center justify-center bg-muted/40 p-4">
      <div className="flex w-full max-w-md flex-col items-center gap-6 text-center">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src="/logo.svg"
          alt={`${APP_NAME} logo`}
          width={64}
          height={64}
        />

        {status === "requested" ? (
          <>
            <CheckCircle2 className="h-12 w-12 text-success" />
            <h1 className="text-2xl font-bold">You&apos;re on the waitlist!</h1>
            <p className="text-muted-foreground">
              We&apos;ll notify you when your access is approved.
            </p>
          </>
        ) : (
          <>
            <ShieldX className="h-12 w-12 text-muted-foreground" />
            <h1 className="text-2xl font-bold">Invite Required</h1>
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
                {status === "loading" && (
                  <Spinner className="mr-2" />
                )}
                Request Access
              </Button>
            )}
          </>
        )}

        <Button
          variant="outline"
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
    </div>
  );
}
