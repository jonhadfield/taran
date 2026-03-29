"use client";

import { useState } from "react";
import { authClient } from "@/lib/auth-client";
import { apiPost } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { APP_NAME } from "@/lib/config";
import { ShieldX, CheckCircle2, Loader2 } from "lucide-react";

export default function NotInvitedPage() {
  const [status, setStatus] = useState<"idle" | "loading" | "requested">(
    "idle"
  );
  const [error, setError] = useState<string | null>(null);

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
            <CheckCircle2 className="h-12 w-12 text-green-500" />
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
              {APP_NAME} is currently invite-only. Request access below and
              we&apos;ll review your request.
            </p>
            {error && <p className="text-sm text-red-500">{error}</p>}
            <Button
              onClick={handleRequestAccess}
              disabled={status === "loading"}
            >
              {status === "loading" && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              Request Access
            </Button>
          </>
        )}

        <Button
          variant="outline"
          onClick={async () => {
            await authClient.signOut();
            window.location.href = "/login";
          }}
        >
          Sign out
        </Button>
      </div>
    </div>
  );
}
