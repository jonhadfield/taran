"use client";

import { useState, useEffect, useCallback } from "react";
import { apiGet, apiPatch } from "@/lib/api";
import type { SenderInfo, SenderSuggestion } from "@/types/api";
import { X, MailX } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import { toast } from "sonner";
import { pluralize } from "@/lib/utils";
import { NativeSelect } from "@/components/ui/native-select";
import { SendersIllustration } from "@/components/empty-state";
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty";
import Link from "next/link";
import { SenderSparkline } from "./sender-sparkline";
import { SenderAvatar } from "@/components/sender-avatar";
import { STATUS_OPTIONS, CATEGORY_OPTIONS, CATEGORY_COLORS, ALL_CATEGORIES } from "@/lib/category-constants";

export default function SendersPage() {
  const [senders, setSenders] = useState<SenderInfo[]>([]);
  const [suggestions, setSuggestions] = useState<SenderSuggestion[]>([]);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<string | null>(null);
  const [dismissedSuggestions, setDismissedSuggestions] = useState(false);
  const [categoryFilter, setCategoryFilter] = useState<string>("all");

  const fetchSenders = useCallback(async () => {
    try {
      const data = await apiGet<SenderInfo[]>("senders");
      setSenders(
        (data || []).sort((a, b) => b.EmailCount - a.EmailCount)
      );
    } catch {
      // keep existing data
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchSuggestions = useCallback(async () => {
    try {
      const data = await apiGet<SenderSuggestion[]>("senders/suggestions");
      setSuggestions(data || []);
    } catch {
      // ignore
    }
  }, []);

  useEffect(() => {
    fetchSenders();
    fetchSuggestions();
  }, [fetchSenders, fetchSuggestions]);

  const handleUpdate = async (fromAddress: string, updates: { Status?: string; Category?: string }) => {
    setUpdating(fromAddress);
    try {
      const sender = senders.find((s) => s.FromAddress === fromAddress);
      await apiPatch("senders", {
        FromAddress: fromAddress,
        Status: updates.Status ?? sender?.Status ?? "normal",
        Category: updates.Category ?? (sender?.Category !== sender?.AutoCategory ? sender?.Category : "") ?? "",
      });
      setSenders((prev) =>
        prev.map((s) => {
          if (s.FromAddress !== fromAddress) return s;
          return {
            ...s,
            ...(updates.Status !== undefined && { Status: updates.Status }),
            ...(updates.Category !== undefined && { Category: updates.Category || s.AutoCategory }),
          };
        })
      );
    } catch {
      // keep existing state
    } finally {
      setUpdating(null);
    }
  };

  const handleMuteSuggestion = async (fromAddress: string) => {
    setUpdating(fromAddress);
    try {
      await apiPatch("senders", { FromAddress: fromAddress, Status: "muted" });
      setSuggestions((prev) => prev.filter((s) => s.FromAddress !== fromAddress));
      setSenders((prev) =>
        prev.map((s) =>
          s.FromAddress === fromAddress ? { ...s, Status: "muted" } : s
        )
      );
    } catch {
      // keep existing state
    } finally {
      setUpdating(null);
    }
  };

  const [mutingAll, setMutingAll] = useState(false);

  const handleMuteAll = async () => {
    setMutingAll(true);
    try {
      const results = await Promise.allSettled(
        suggestions.map((s) =>
          apiPatch("senders", { FromAddress: s.FromAddress, Status: "muted" }).then(() => s.FromAddress)
        )
      );
      const succeeded = new Set(
        results
          .filter((r): r is PromiseFulfilledResult<string> => r.status === "fulfilled")
          .map((r) => r.value)
      );
      const failedCount = results.length - succeeded.size;
      setSenders((prev) =>
        prev.map((s) => {
          if (succeeded.has(s.FromAddress)) {
            return { ...s, Status: "muted" };
          }
          return s;
        })
      );
      setSuggestions((prev) => prev.filter((s) => !succeeded.has(s.FromAddress)));
      if (failedCount > 0) {
        toast.error(`Failed to mute ${failedCount} ${pluralize(failedCount, "sender")}`);
      }
    } finally {
      setMutingAll(false);
    }
  };

  const filteredSenders = categoryFilter === "all"
    ? senders
    : senders.filter((s) => s.Category === categoryFilter);

  // Count senders per category
  const categoryCounts = senders.reduce<Record<string, number>>((acc, s) => {
    const cat = s.Category || "other";
    acc[cat] = (acc[cat] || 0) + 1;
    return acc;
  }, {});

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-balance">Senders</h1>
        <div className="flex items-center gap-3">
          <Link
            href="/senders/subscriptions"
            className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            <MailX className="size-4" />
            Manage Subscriptions
          </Link>
          <span className="text-sm text-muted-foreground">
            {filteredSenders.length} of {senders.length} senders
          </span>
        </div>
      </div>

      {/* Category filter tabs */}
      {!loading && senders.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          <button
            onClick={() => setCategoryFilter("all")}
            className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
              categoryFilter === "all"
                ? "bg-foreground text-background"
                : "bg-muted text-muted-foreground hover:bg-accent"
            }`}
          >
            All ({senders.length})
          </button>
          {ALL_CATEGORIES.filter((cat) => categoryCounts[cat]).map((cat) => (
            <button
              key={cat}
              onClick={() => setCategoryFilter(cat)}
              className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                categoryFilter === cat
                  ? "bg-foreground text-background"
                  : CATEGORY_COLORS[cat] + " hover:opacity-80"
              }`}
            >
              {cat.charAt(0).toUpperCase() + cat.slice(1)} ({categoryCounts[cat]})
            </button>
          ))}
        </div>
      )}

      {/* Mute suggestions banner */}
      {suggestions.length > 0 && !dismissedSuggestions && (
        <div className="rounded-lg border border-amber-200 bg-amber-50 dark:border-amber-900 dark:bg-amber-950/30 p-4">
          <div className="flex items-start justify-between gap-2">
            <div className="space-y-2">
              <p className="text-sm font-medium">
                Based on your feedback, you may want to mute these senders:
              </p>
              <div className="space-y-1.5">
                {suggestions.map((s) => (
                  <div
                    key={s.FromAddress}
                    className="flex items-center gap-2 text-sm"
                  >
                    <span className="truncate">
                      {s.FromName || s.FromAddress}
                    </span>
                    <span className="text-xs text-muted-foreground shrink-0">
                      ({s.NotUsefulCount}/{s.TotalCount} not useful)
                    </span>
                    <button
                      onClick={() => handleMuteSuggestion(s.FromAddress)}
                      disabled={updating === s.FromAddress}
                      className="shrink-0 rounded-md border px-2 py-0.5 text-xs hover:bg-accent transition-colors disabled:opacity-50"
                    >
                      Mute
                    </button>
                  </div>
                ))}
              </div>
              {suggestions.length > 1 && (
                <button
                  onClick={handleMuteAll}
                  disabled={mutingAll}
                  className="mt-1 text-xs font-medium text-amber-700 dark:text-amber-400 hover:underline disabled:opacity-50"
                >
                  {mutingAll ? "Muting..." : `Mute all ${suggestions.length} senders`}
                </button>
              )}
            </div>
            <button
              onClick={() => setDismissedSuggestions(true)}
              className="shrink-0 text-muted-foreground hover:text-foreground"
            >
              <X className="size-4" />
            </button>
          </div>
        </div>
      )}

      {loading ? (
        <div className="flex items-center gap-2 py-12 justify-center text-muted-foreground">
          <Spinner className="size-5" />
          Loading senders...
        </div>
      ) : senders.length === 0 ? (
        <Empty>
          <EmptyHeader>
            <EmptyMedia><SendersIllustration /></EmptyMedia>
            <EmptyTitle>No senders yet</EmptyTitle>
            <EmptyDescription>Senders will appear here once you receive emails.</EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="divide-y rounded-lg border">
          {filteredSenders.map((sender) => {
            const name = sender.FromName || sender.FromAddress;
            const isOverridden = sender.Category !== sender.AutoCategory && sender.AutoCategory !== "";

            return (
              <div
                key={sender.FromAddress}
                className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3 p-3.5"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <SenderAvatar name={name} />

                  <Link
                    href={`/senders/${encodeURIComponent(sender.FromAddress)}`}
                    className="flex-1 min-w-0 hover:opacity-80 transition-opacity"
                  >
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium truncate">{name}</p>
                      {sender.Category && (
                        <span className={`shrink-0 rounded-full px-2 py-0.5 text-[10px] font-medium ${CATEGORY_COLORS[sender.Category] || CATEGORY_COLORS.other}`}>
                          {sender.Category}
                          {isOverridden && " *"}
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground truncate">
                      {sender.FromAddress}
                    </p>
                  </Link>

                  <span className="text-xs text-muted-foreground whitespace-nowrap shrink-0">
                    {sender.EmailCount} {pluralize(sender.EmailCount, "email")}
                  </span>
                </div>

                <div className="flex items-center gap-2 sm:shrink-0">
                  <div className="hidden sm:block">
                    <SenderSparkline fromAddress={sender.FromAddress} />
                  </div>

                  <NativeSelect
                    size="sm"
                    value={sender.Category === sender.AutoCategory ? "" : sender.Category}
                    onChange={(e) =>
                      handleUpdate(sender.FromAddress, { Category: e.target.value })
                    }
                    disabled={updating === sender.FromAddress}
                    wrapperClassName="flex-1 sm:flex-none sm:w-[100px]"
                    className="pl-2 text-xs"
                  >
                    {CATEGORY_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.value === "" && sender.AutoCategory
                          ? `Auto (${sender.AutoCategory})`
                          : opt.label}
                      </option>
                    ))}
                  </NativeSelect>

                  <NativeSelect
                    size="sm"
                    value={sender.Status}
                    onChange={(e) =>
                      handleUpdate(sender.FromAddress, { Status: e.target.value })
                    }
                    disabled={updating === sender.FromAddress}
                    wrapperClassName="flex-1 sm:flex-none sm:w-[120px]"
                    className="pl-2 text-xs"
                  >
                    {STATUS_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </NativeSelect>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
