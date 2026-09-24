"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { apiGet, apiPost, apiPatch, apiDelete, ApiError } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { NativeSelect } from "@/components/ui/native-select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Pencil, Trash2, Plus, RefreshCw } from "lucide-react";
import type { AnalysisRule } from "@/types/api";
import {
  ANALYSIS_RULE_EXAMPLES,
  MAX_ANALYSIS_RULES,
  MAX_ANALYSIS_RULE_LENGTH,
  MAX_REANALYSE_EMAILS,
  REANALYSE_DAY_OPTIONS,
} from "@/lib/analysis-rules";
import { pluralize } from "@/lib/utils";

function errorMessage(err: unknown, fallback: string) {
  return err instanceof ApiError && err.status === 400 ? err.message : fallback;
}

export function AnalysisRulesSettings() {
  const [rules, setRules] = useState<AnalysisRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [draft, setDraft] = useState("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editText, setEditText] = useState("");
  const [showApply, setShowApply] = useState(false);
  const [applyDays, setApplyDays] = useState(7);
  const [applying, setApplying] = useState(false);

  useEffect(() => {
    apiGet<AnalysisRule[]>("analysis-rules")
      .then((data) => setRules(data || []))
      .catch(() => {
        // keep existing
      })
      .finally(() => setLoading(false));
  }, []);

  const atLimit = rules.length >= MAX_ANALYSIS_RULES;

  const handleAdd = async () => {
    setSaving(true);
    try {
      const updated = await apiPost<AnalysisRule[]>("analysis-rules", {
        Rule: draft.trim(),
      });
      setRules(updated || []);
      setDraft("");
      setShowForm(false);
      toast.success("Rule added");
    } catch (err) {
      toast.error(errorMessage(err, "Failed to save rule"));
    } finally {
      setSaving(false);
    }
  };

  const handleUpdate = async (
    id: string,
    patch: { Rule?: string; IsActive?: boolean },
  ) => {
    try {
      const updated = await apiPatch<AnalysisRule[]>(
        `analysis-rules/${id}`,
        patch,
      );
      setRules(updated || []);
      return true;
    } catch (err) {
      toast.error(errorMessage(err, "Failed to update rule"));
      return false;
    }
  };

  const handleSaveEdit = async (id: string) => {
    setSaving(true);
    if (await handleUpdate(id, { Rule: editText.trim() })) {
      setEditingId(null);
      toast.success("Rule updated");
    }
    setSaving(false);
  };

  const handleDelete = async (id: string) => {
    try {
      await apiDelete(`analysis-rules/${id}`);
      setRules((prev) => prev.filter((r) => r.ID !== id));
      toast.success("Rule removed");
    } catch {
      toast.error("Failed to delete rule");
    }
  };

  const handleApply = async () => {
    setApplying(true);
    try {
      const { queued, matched } = await apiPost<{ queued: number; matched: number }>(
        "emails/reanalyse",
        { Days: applyDays },
      );
      setShowApply(false);
      if (queued === 0) {
        toast.info("No processed emails in that period");
      } else {
        const more =
          matched > queued ? ` (the ${queued} most recent of ${matched})` : "";
        toast.success(
          `Re-analysing ${queued} ${pluralize(queued, "email")}${more}. Summaries update as each one finishes.`,
        );
      }
    } catch (err) {
      toast.error(errorMessage(err, "Failed to start re-analysis"));
    } finally {
      setApplying(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Analysis rules</CardTitle>
        <CardDescription>
          Tell the AI how to analyse your emails. Rules apply to newly received
          emails and to your digests, and you can apply them to emails you
          have already received.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {loading ? (
          <p className="text-sm text-muted-foreground">Loading...</p>
        ) : (
          <>
            {rules.length > 0 && (
              <div className="divide-y rounded-lg border">
                {rules.map((rule) =>
                  editingId === rule.ID ? (
                    <div key={rule.ID} className="space-y-2 px-3 py-2.5">
                      <Textarea
                        value={editText}
                        onChange={(e) => setEditText(e.target.value)}
                        maxLength={MAX_ANALYSIS_RULE_LENGTH}
                        aria-label="Edit rule"
                        autoFocus
                      />
                      <div className="flex items-center gap-2">
                        <Button
                          size="sm"
                          onClick={() => handleSaveEdit(rule.ID)}
                          disabled={saving || !editText.trim()}
                        >
                          Save
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => setEditingId(null)}
                        >
                          Cancel
                        </Button>
                        <span className="ml-auto text-xs text-muted-foreground">
                          {editText.length}/{MAX_ANALYSIS_RULE_LENGTH}
                        </span>
                      </div>
                    </div>
                  ) : (
                    <div
                      key={rule.ID}
                      className="flex items-start justify-between gap-3 px-3 py-2.5"
                    >
                      <p
                        className={`min-w-0 flex-1 text-sm break-words ${
                          rule.IsActive ? "" : "text-muted-foreground line-through"
                        }`}
                      >
                        {rule.Rule}
                      </p>
                      <div className="flex shrink-0 items-center gap-3">
                        <Switch
                          size="sm"
                          checked={rule.IsActive}
                          onCheckedChange={(checked) =>
                            handleUpdate(rule.ID, { IsActive: checked })
                          }
                          aria-label={rule.IsActive ? "Disable rule" : "Enable rule"}
                        />
                        <button
                          onClick={() => {
                            setEditingId(rule.ID);
                            setEditText(rule.Rule);
                          }}
                          className="text-muted-foreground hover:text-foreground transition-colors"
                          aria-label="Edit rule"
                        >
                          <Pencil className="size-4" />
                        </button>
                        <button
                          onClick={() => handleDelete(rule.ID)}
                          className="text-muted-foreground hover:text-destructive transition-colors"
                          aria-label="Delete rule"
                        >
                          <Trash2 className="size-4" />
                        </button>
                      </div>
                    </div>
                  ),
                )}
              </div>
            )}

            {rules.length === 0 && !showForm && (
              <div className="space-y-1.5 text-sm text-muted-foreground">
                <p>No rules yet. For example:</p>
                <ul className="list-disc space-y-0.5 pl-5">
                  {ANALYSIS_RULE_EXAMPLES.map((example) => (
                    <li key={example}>{example}</li>
                  ))}
                </ul>
              </div>
            )}

            {showForm ? (
              <div className="space-y-2 rounded-lg border p-3">
                <Label htmlFor="new-analysis-rule">New rule</Label>
                <Textarea
                  id="new-analysis-rule"
                  value={draft}
                  onChange={(e) => setDraft(e.target.value)}
                  maxLength={MAX_ANALYSIS_RULE_LENGTH}
                  placeholder={ANALYSIS_RULE_EXAMPLES[0]}
                  autoFocus
                />
                <div className="flex items-center gap-2 pt-1">
                  <Button
                    size="sm"
                    onClick={handleAdd}
                    disabled={saving || !draft.trim()}
                  >
                    Save
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => {
                      setShowForm(false);
                      setDraft("");
                    }}
                  >
                    Cancel
                  </Button>
                  <span className="ml-auto text-xs text-muted-foreground">
                    {draft.length}/{MAX_ANALYSIS_RULE_LENGTH}
                  </span>
                </div>
              </div>
            ) : (
              <div className="flex flex-wrap items-center gap-3">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowForm(true)}
                  disabled={atLimit}
                  className="gap-1.5"
                >
                  <Plus className="size-3.5" />
                  Add rule
                </Button>
                {rules.length > 0 && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowApply(true)}
                    className="gap-1.5"
                  >
                    <RefreshCw className="size-3.5" />
                    Apply to recent emails
                  </Button>
                )}
                <span className="text-xs text-muted-foreground">
                  {rules.length}/{MAX_ANALYSIS_RULES} rules
                </span>
              </div>
            )}
          </>
        )}
      </CardContent>

      <Dialog open={showApply} onOpenChange={setShowApply}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Apply rules to recent emails</DialogTitle>
            <DialogDescription>
              Re-analyse your processed emails with your current rules. Up to{" "}
              {MAX_REANALYSE_EMAILS} of the most
              recent emails are re-analysed, which uses AI tokens. If an email
              can&apos;t be re-analysed, its previous summary is kept.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label htmlFor="reanalyse-days">Emails received in the last</Label>
            <NativeSelect
              id="reanalyse-days"
              value={String(applyDays)}
              onChange={(e) => setApplyDays(Number(e.target.value))}
              wrapperClassName="w-full"
            >
              {REANALYSE_DAY_OPTIONS.map((days) => (
                <option key={days} value={days}>
                  {days} {pluralize(days, "day")}
                </option>
              ))}
            </NativeSelect>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setShowApply(false)}>
              Cancel
            </Button>
            <Button onClick={handleApply} disabled={applying}>
              Re-analyse
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
