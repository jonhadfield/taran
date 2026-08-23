"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { AlertCircle } from "lucide-react";

export default function DashboardError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12 text-center">
        <AlertCircle className="h-12 w-12 text-destructive mb-4" />
        <h3 className="text-lg font-medium">Something went wrong</h3>
        <p className="text-sm text-muted-foreground mt-1 mb-4">
          Failed to load this page. Please try again.
        </p>
        <div className="flex gap-2">
          <Button onClick={reset}>Try Again</Button>
          {/*
            Full document load on purpose. "Try Again" above is the soft path
            (reset() re-renders the segment); this button is the escape hatch
            for when the client state itself is what broke, so it deliberately
            starts a fresh document rather than navigating within it.
          */}
          <Button
            variant="outline"
            // eslint-disable-next-line @next/next/no-location-assign-relative-destination
            onClick={() => (window.location.href = "/")}
          >
            Dashboard
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
