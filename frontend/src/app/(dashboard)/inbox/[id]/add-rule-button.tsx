"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { apiPost, ApiError } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { Wand2 } from "lucide-react";
import type { AnalysisRule } from "@/types/api";
import { MAX_ANALYSIS_RULE_LENGTH } from "@/lib/analysis-rules";
import { reanalyseEmail, reanalysisMessage } from "@/lib/reanalyse";

/** Builds an editable starting point for a rule from what the AI found in an email. */
export function suggestRule(topics: string[], senderName: string): string {
  const cleaned = topics.map((t) => t.trim()).filter(Boolean);
  if (cleaned.length > 0) {
    const list =
      cleaned.length === 1
        ? cleaned[0]
        : `${cleaned.slice(0, -1).join(", ")} and ${cleaned[cleaned.length - 1]}`;
    return `I want more detail on emails about ${list}`;
  }
  if (senderName.trim()) {
    return `I want more detail on emails from ${senderName.trim()}`;
  }
  return "";
}

export function AddRuleButton({
  emailId,
  processedAt,
  topics,
  senderName,
}: {
  emailId: string;
  processedAt: string;
  topics: string[];
  senderName: string;
}) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [rule, setRule] = useState("");
  const [reanalyse, setReanalyse] = useState(true);
  const [saving, setSaving] = useState(false);

  const handleOpenChange = (next: boolean) => {
    if (next) {
      setRule(suggestRule(topics, senderName).slice(0, MAX_ANALYSIS_RULE_LENGTH));
      setReanalyse(true);
    }
    setOpen(next);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await apiPost<AnalysisRule[]>("analysis-rules", { Rule: rule.trim() });
    } catch (err) {
      toast.error(
        err instanceof ApiError && err.status === 400
          ? err.message
          : "Failed to save rule",
      );
      setSaving(false);
      return;
    }

    setOpen(false);
    setSaving(false);
    toast.success("Analysis rule added", {
      action: {
        label: "Manage rules",
        onClick: () => router.push("/settings#analysis-rules"),
      },
    });
    if (!reanalyse) return;

    const pending = toast.loading("Re-analysing this email with your rules...");
    try {
      const { ok, text } = reanalysisMessage(await reanalyseEmail(emailId, processedAt));
      if (ok) {
        toast.success(text, { id: pending });
      } else {
        toast.error(text, { id: pending });
      }
      router.refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to re-analyse", { id: pending });
    }
  };

  return (
    <>
      <Button
        variant="ghost"
        size="sm"
        className="gap-1.5 text-muted-foreground"
        onClick={() => handleOpenChange(true)}
      >
        <Wand2 className="size-4" />
        Add analysis rule
      </Button>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add analysis rule</DialogTitle>
            <DialogDescription>
              Tell the AI how to handle emails like this one. The rule applies
              to future emails and digests, and you can re-analyse existing
              emails with it.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-1">
            <Textarea
              value={rule}
              onChange={(e) => setRule(e.target.value)}
              maxLength={MAX_ANALYSIS_RULE_LENGTH}
              aria-label="Analysis rule"
              autoFocus
            />
            <p className="text-right text-xs text-muted-foreground">
              {rule.length}/{MAX_ANALYSIS_RULE_LENGTH}
            </p>
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={reanalyse}
              onChange={(e) => setReanalyse(e.target.checked)}
              className="size-4 accent-primary"
            />
            Re-analyse this email now with the new rule
          </label>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={saving || !rule.trim()}>
              Save rule
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
