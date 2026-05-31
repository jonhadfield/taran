"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { apiGet, apiPut, apiDelete } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Trash2, Plus } from "lucide-react";
import type { AutoArchiveRule } from "@/types/api";
import { pluralize } from "@/lib/utils";
import { NativeSelect } from "@/components/ui/native-select";
import { CATEGORY_OPTIONS_NO_AUTO } from "@/lib/category-constants";

export function AutoArchiveSettings() {
  const [rules, setRules] = useState<AutoArchiveRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showForm, setShowForm] = useState(false);

  // Form state
  const [ruleType, setRuleType] = useState<"category" | "sender">("category");
  const [ruleValue, setRuleValue] = useState("newsletter");
  const [archiveDays, setArchiveDays] = useState(7);

  const fetchRules = async () => {
    try {
      const data = await apiGet<AutoArchiveRule[]>("auto-archive-rules");
      setRules(data || []);
    } catch {
      // keep existing
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRules();
  }, []);

  const handleAdd = async () => {
    setSaving(true);
    try {
      const updated = await apiPut<AutoArchiveRule[]>("auto-archive-rules", {
        RuleType: ruleType,
        RuleValue: ruleValue,
        ArchiveAfterDays: archiveDays,
      });
      setRules(updated || []);
      setShowForm(false);
      setRuleValue(ruleType === "category" ? "newsletter" : "");
      setArchiveDays(7);
      toast.success("Rule added");
    } catch {
      toast.error("Failed to save rule");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await apiDelete(`auto-archive-rules/${id}`);
      setRules((prev) => prev.filter((r) => r.ID !== id));
      toast.success("Rule removed");
    } catch {
      toast.error("Failed to delete rule");
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Auto-Archive Rules</CardTitle>
        <CardDescription>
          Automatically archive old emails by category or sender
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {loading ? (
          <p className="text-sm text-muted-foreground">Loading...</p>
        ) : (
          <>
            {rules.length > 0 && (
              <div className="divide-y rounded-lg border">
                {rules.map((rule) => (
                  <div
                    key={rule.ID}
                    className="flex items-center justify-between gap-3 px-3 py-2.5"
                  >
                    <div className="text-sm">
                      <span className="font-medium capitalize">{rule.RuleType}:</span>{" "}
                      <span>{rule.RuleValue}</span>
                      <span className="text-muted-foreground">
                        {" "}— after {rule.ArchiveAfterDays} {pluralize(rule.ArchiveAfterDays, "day")}
                      </span>
                    </div>
                    <button
                      onClick={() => handleDelete(rule.ID)}
                      className="shrink-0 text-muted-foreground hover:text-destructive transition-colors"
                      aria-label="Delete rule"
                    >
                      <Trash2 className="size-4" />
                    </button>
                  </div>
                ))}
              </div>
            )}

            {rules.length === 0 && !showForm && (
              <p className="text-sm text-muted-foreground">
                No rules configured. Add a rule to automatically archive older emails.
              </p>
            )}

            {showForm ? (
              <div className="space-y-3 rounded-lg border p-3">
                <div className="space-y-2">
                  <Label>Rule type</Label>
                  <div className="flex gap-2">
                    <Button
                      type="button"
                      size="sm"
                      variant={ruleType === "category" ? "default" : "outline"}
                      onClick={() => {
                        setRuleType("category");
                        setRuleValue("newsletter");
                      }}
                    >
                      Category
                    </Button>
                    <Button
                      type="button"
                      size="sm"
                      variant={ruleType === "sender" ? "default" : "outline"}
                      onClick={() => {
                        setRuleType("sender");
                        setRuleValue("");
                      }}
                    >
                      Sender
                    </Button>
                  </div>
                </div>

                <div className="space-y-2">
                  <Label>{ruleType === "category" ? "Category" : "Sender email"}</Label>
                  {ruleType === "category" ? (
                    <NativeSelect
                      value={ruleValue}
                      onChange={(e) => setRuleValue(e.target.value)}
                      wrapperClassName="w-full"
                    >
                      {CATEGORY_OPTIONS_NO_AUTO.map((opt) => (
                        <option key={opt.value} value={opt.value}>
                          {opt.label}
                        </option>
                      ))}
                    </NativeSelect>
                  ) : (
                    <Input
                      type="email"
                      value={ruleValue}
                      onChange={(e) => setRuleValue(e.target.value)}
                      placeholder="sender@example.com"
                    />
                  )}
                </div>

                <div className="space-y-2">
                  <Label>Archive after (days)</Label>
                  <Input
                    type="number"
                    min={1}
                    max={365}
                    value={archiveDays}
                    onChange={(e) => setArchiveDays(Number(e.target.value))}
                    className="w-24"
                  />
                </div>

                <div className="flex gap-2 pt-1">
                  <Button
                    size="sm"
                    onClick={handleAdd}
                    disabled={saving || !ruleValue}
                  >
                    Save
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => setShowForm(false)}
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            ) : (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowForm(true)}
                className="gap-1.5"
              >
                <Plus className="size-3.5" />
                Add rule
              </Button>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}
