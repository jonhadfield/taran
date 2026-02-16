"use client";

import { authClient } from "@/lib/auth-client";
import { Button } from "@/components/ui/button";
import { APP_NAME } from "@/lib/config";
import { ShieldX } from "lucide-react";

export default function NotInvitedPage() {
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
        <ShieldX className="h-12 w-12 text-muted-foreground" />
        <h1 className="text-2xl font-bold">Invite Required</h1>
        <p className="text-muted-foreground">
          {APP_NAME} is currently invite-only. You need an invitation from an
          existing member to access the dashboard.
        </p>
        <p className="text-sm text-muted-foreground">
          If you believe you should have access, please contact the person who
          told you about {APP_NAME}.
        </p>
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
