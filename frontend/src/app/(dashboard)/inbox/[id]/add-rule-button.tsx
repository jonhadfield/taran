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
  topics,
  senderName,
}: {
  topics: string[];
  senderName: string;
}) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [rule, setRule] = useState("");
  const [saving, setSaving] = useState(false);

  const handleOpenChange = (next: boolean) => {
    if (next) {
      setRule(suggestRule(topics, senderName).slice(0, MAX_ANALYSIS_RULE_LENGTH));
    }
    setOpen(next);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await apiPost<AnalysisRule[]>("analysis-rules", { Rule: rule.trim() });
      setOpen(false);
      toast.success("Analysis rule added", {
        action: {
          label: "Manage rules",
          onClick: () => router.push("/settings#analysis-rules"),
        },
      });
    } catch (err) {
      toast.error(
        err instanceof ApiError && err.status === 400
          ? err.message
          : "Failed to save rule",
      );
    } finally {
      setSaving(false);
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
              to future emails and digests.
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
