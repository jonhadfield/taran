"use client";

import * as Sentry from "@sentry/nextjs";
import { useEffect } from "react";
import { Button } from "@/components/ui/button";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    Sentry.captureException(error);
  }, [error]);

  return (
    <html lang="en">
      <body className="antialiased bg-background text-foreground">
        <div className="min-h-screen flex items-center justify-center px-4">
          <div className="text-center max-w-md space-y-4">
            <h1 className="text-6xl font-bold text-muted-foreground/40">500</h1>
            <h2 className="text-xl font-semibold">Something went wrong</h2>
            <p className="text-sm text-muted-foreground">
              {error.message || "A critical error occurred."}
            </p>
            {error.digest && (
              <p className="text-xs text-muted-foreground">Error ID: {error.digest}</p>
            )}
            <Button onClick={reset}>Try again</Button>
          </div>
        </div>
      </body>
    </html>
  );
}
