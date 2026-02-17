"use client";

import { useState, useEffect } from "react";
import { apiGet, apiPatch } from "@/lib/api";
import type { SenderInfo, SenderSuggestion } from "@/types/api";
import { Users, Loader2, X } from "lucide-react";

const avatarColors = [
  "bg-blue-500",
  "bg-emerald-500",
  "bg-amber-500",
  "bg-rose-500",
  "bg-purple-500",
  "bg-cyan-500",
  "bg-indigo-500",
  "bg-pink-500",
];

function getAvatarColor(name: string) {
  const charCode = name.charCodeAt(0) || 0;
  return avatarColors[charCode % avatarColors.length];
}

const STATUS_OPTIONS = [
  { value: "normal", label: "Normal" },
  { value: "favorite", label: "Favorite" },
  { value: "muted", label: "Muted" },
  { value: "blocked", label: "Blocked" },
];

export default function SendersPage() {
  const [senders, setSenders] = useState<SenderInfo[]>([]);
  const [suggestions, setSuggestions] = useState<SenderSuggestion[]>([]);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<string | null>(null);
  const [dismissedSuggestions, setDismissedSuggestions] = useState(false);

  const fetchSenders = async () => {
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
  };

  const fetchSuggestions = async () => {
    try {
      const data = await apiGet<SenderSuggestion[]>("senders/suggestions");
      setSuggestions(data || []);
    } catch {
      // ignore
    }
  };

  useEffect(() => {
    fetchSenders();
    fetchSuggestions();
  }, []);

  const handleStatusChange = async (fromAddress: string, status: string) => {
    setUpdating(fromAddress);
    try {
      await apiPatch("senders", { FromAddress: fromAddress, Status: status });
      setSenders((prev) =>
        prev.map((s) =>
          s.FromAddress === fromAddress ? { ...s, Status: status } : s
        )
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

  const handleMuteAll = async () => {
    for (const s of suggestions) {
      await apiPatch("senders", { FromAddress: s.FromAddress, Status: "muted" });
    }
    setSenders((prev) =>
      prev.map((s) => {
        if (suggestions.some((sg) => sg.FromAddress === s.FromAddress)) {
          return { ...s, Status: "muted" };
        }
        return s;
      })
    );
    setSuggestions([]);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-balance">Senders</h1>
        <span className="text-sm text-muted-foreground">
          {senders.length} senders
        </span>
      </div>

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
                  className="mt-1 text-xs font-medium text-amber-700 dark:text-amber-400 hover:underline"
                >
                  Mute all {suggestions.length} senders
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
          <Loader2 className="size-5 animate-spin" />
          Loading senders...
        </div>
      ) : senders.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Users className="size-12 text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium">No senders yet</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Senders will appear here once you receive emails.
          </p>
        </div>
      ) : (
        <div className="divide-y rounded-lg border">
          {senders.map((sender) => {
            const name = sender.FromName || sender.FromAddress;
            const initial = name.charAt(0).toUpperCase();

            return (
              <div
                key={sender.FromAddress}
                className="flex items-center gap-3 p-3.5"
              >
                <div
                  className={`hidden sm:flex size-9 shrink-0 items-center justify-center rounded-full text-sm font-medium text-white ${getAvatarColor(name)}`}
                >
                  {initial}
                </div>

                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{name}</p>
                  <p className="text-xs text-muted-foreground truncate">
                    {sender.FromAddress}
                  </p>
                </div>

                <span className="text-xs text-muted-foreground whitespace-nowrap">
                  {sender.EmailCount} email{sender.EmailCount !== 1 ? "s" : ""}
                </span>

                <select
                  value={sender.Status}
                  onChange={(e) =>
                    handleStatusChange(sender.FromAddress, e.target.value)
                  }
                  disabled={updating === sender.FromAddress}
                  className="h-8 rounded-md border border-input bg-transparent px-2 text-xs shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring max-w-full sm:max-w-[120px]"
                >
                  {STATUS_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
