"use client";

import { useState, useEffect } from "react";
import { apiGet, apiPatch } from "@/lib/api";
import type { SenderInfo } from "@/types/api";
import { Users, Loader2 } from "lucide-react";

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
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<string | null>(null);

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

  useEffect(() => {
    fetchSenders();
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

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-balance">Senders</h1>
        <span className="text-sm text-muted-foreground">
          {senders.length} senders
        </span>
      </div>

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
