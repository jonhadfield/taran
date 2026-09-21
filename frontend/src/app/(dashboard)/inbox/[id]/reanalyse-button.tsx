"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { RefreshCw } from "lucide-react";
import { reanalyseEmail, reanalysisMessage } from "@/lib/reanalyse";

/** Re-runs AI analysis on a processed email, e.g. after changing analysis rules. */
export function ReanalyseButton({
  emailId,
  processedAt,
}: {
  emailId: string;
  processedAt: string;
}) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);

  async function handleClick() {
    setBusy(true);
    try {
      const outcome = await reanalyseEmail(emailId, processedAt);
      const { ok, text } = reanalysisMessage(outcome);
      if (ok) {
        toast.success(text);
      } else {
        toast.error(text);
      }
      router.refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to re-analyse");
    } finally {
      setBusy(false);
    }
  }

  const label = busy ? "Re-analysing..." : "Re-analyse";

  return (
    <Button
      variant="ghost"
      size="sm"
      className="gap-1.5 text-muted-foreground"
      onClick={handleClick}
      disabled={busy}
      aria-label={label}
      title={label}
    >
      <RefreshCw className={`size-4 ${busy ? "animate-spin" : ""}`} />
      {/* Icon-only on narrow screens so the card title keeps its line */}
      <span className="hidden sm:inline">{label}</span>
    </Button>
  );
}
