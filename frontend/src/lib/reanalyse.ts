import { apiGet, apiPost } from "@/lib/api";
import type { EmailResponse } from "@/types/api";

export type ReanalysisOutcome = "updated" | "unchanged" | "failed" | "timeout";

/**
 * Queues an email for AI re-analysis and waits for the worker to finish.
 *
 * A processed email keeps its previous summary if re-analysis fails, and
 * returns to "processed" either way, so success is detected by the
 * extraction's ProcessedAt changing rather than by status alone.
 */
export async function reanalyseEmail(
  emailId: string,
  previousProcessedAt: string | undefined,
  { intervalMs = 2000, attempts = 30 } = {},
): Promise<ReanalysisOutcome> {
  await apiPost(`emails/${emailId}/reprocess`, {});

  for (let i = 0; i < attempts; i++) {
    await new Promise((r) => setTimeout(r, intervalMs));
    try {
      const email = await apiGet<EmailResponse>(`emails/${emailId}`);
      if (email.Status === "processed") {
        return email.Extraction?.ProcessedAt !== previousProcessedAt
          ? "updated"
          : "unchanged";
      }
      if (email.Status === "failed" || email.Status === "skipped") {
        return "failed";
      }
    } catch {
      // Poll failed, keep trying
    }
  }
  return "timeout";
}

/** Toast-ready message for a re-analysis outcome. */
export function reanalysisMessage(outcome: ReanalysisOutcome): {
  ok: boolean;
  text: string;
} {
  switch (outcome) {
    case "updated":
      return { ok: true, text: "Summary updated" };
    case "unchanged":
    case "failed":
      return { ok: false, text: "Re-analysis failed, so the previous summary was kept" };
    case "timeout":
      return { ok: true, text: "Still re-analysing, so refresh in a moment" };
  }
}
