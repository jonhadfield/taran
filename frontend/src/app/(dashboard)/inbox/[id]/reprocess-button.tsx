"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { RefreshCw, CheckCircle2 } from "lucide-react";
import { apiPost, apiGet } from "@/lib/api";
import type { EmailResponse } from "@/types/api";

export function ReprocessButton({ emailId }: { emailId: string }) {
  const router = useRouter();
  const [state, setState] = useState<"idle" | "submitting" | "processing" | "done" | "error">("idle");
  const [error, setError] = useState<string | null>(null);

  async function handleClick() {
    setState("submitting");
    setError(null);
    try {
      await apiPost(`emails/${emailId}/reprocess`, {});
      setState("processing");

      // Poll for completion (up to 30s)
      for (let i = 0; i < 15; i++) {
        await new Promise((r) => setTimeout(r, 2000));
        try {
          const email = await apiGet<EmailResponse>(`emails/${emailId}`);
          if (email.Status === "processed" || email.Status === "skipped") {
            setState("done");
            setTimeout(() => router.refresh(), 1500);
            return;
          }
          if (email.Status === "failed") {
            setState("error");
            setError("Processing failed — check the status reason above");
            return;
          }
        } catch {
          // Poll failed, keep trying
        }
      }

      // Timed out waiting — still show success, refresh page
      setState("done");
      setTimeout(() => router.refresh(), 1500);
    } catch (e) {
      setState("error");
      setError(e instanceof Error ? e.message : "Failed to reprocess");
    }
  }

  if (state === "done") {
    return (
      <div className="mt-2 flex items-center gap-2 text-sm text-success">
        <CheckCircle2 className="size-4" />
        Reprocessed successfully — refreshing...
      </div>
    );
  }

  const isLoading = state === "submitting" || state === "processing";
  const label = state === "submitting"
    ? "Submitting..."
    : state === "processing"
      ? "Processing..."
      : "Reprocess";

  return (
    <div className="mt-2">
      <Button
        variant="outline"
        size="sm"
        onClick={handleClick}
        disabled={isLoading}
      >
        <RefreshCw className={`size-4 mr-1.5 ${isLoading ? "animate-spin" : ""}`} />
        {label}
      </Button>
      {error && <p className="text-sm text-destructive mt-1">{error}</p>}
    </div>
  );
}
