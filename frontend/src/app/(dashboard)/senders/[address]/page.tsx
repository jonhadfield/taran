"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { apiGet, apiPatch } from "@/lib/api";
import type { SenderDetail, WeekCount, Email, ListResponse } from "@/types/api";
import { Loader2, Mail, Calendar, BarChart3, Clock, TrendingUp, ExternalLink } from "lucide-react";
import { formatShortDate } from "@/lib/utils";
import { STATUS_OPTIONS, CATEGORY_OPTIONS, CATEGORY_COLORS } from "@/lib/category-constants";

function VolumeChart({ data }: { data: WeekCount[] }) {
  if (data.length === 0) return null;
  const max = Math.max(...data.map((d) => d.Count), 1);

  return (
    <div className="space-y-2">
      <div className="grid gap-1" style={{ gridTemplateColumns: `repeat(${data.length}, 1fr)`, height: 120 }}>
        {data.map((wc, i) => {
          const pct = Math.max((wc.Count / max) * 100, 2);
          return (
            <div key={i} className="relative flex flex-col justify-end items-center">
              {wc.Count > 0 && (
                <span className="text-[10px] text-muted-foreground mb-1">
                  {wc.Count}
                </span>
              )}
              <div
                className="w-full rounded-t bg-primary/80"
                style={{ height: `${pct}%` }}
              />
            </div>
          );
        })}
      </div>
      <div className="grid gap-1" style={{ gridTemplateColumns: `repeat(${data.length}, 1fr)` }}>
        {data.map((wc, i) => (
          <div key={i} className="text-center">
            <span className="text-[9px] text-muted-foreground">
              {formatShortDate(wc.Week)}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

export default function SenderDetailPage() {
  const params = useParams<{ address: string }>();
  const address = decodeURIComponent(params.address);

  const [detail, setDetail] = useState<SenderDetail | null>(null);
  const [history, setHistory] = useState<WeekCount[]>([]);
  const [emails, setEmails] = useState<Email[]>([]);
  const [emailTotal, setEmailTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);

  useEffect(() => {
    const load = async () => {
      try {
        const [d, h, e] = await Promise.all([
          apiGet<SenderDetail>("senders/detail", { address }),
          apiGet<WeekCount[]>("senders/history", { address, weeks: "12" }),
          apiGet<ListResponse<Email>>("emails", {
            from_address: address,
            limit: "10",
          }),
        ]);
        setDetail(d);
        setHistory(h || []);
        setEmails(e.data || []);
        setEmailTotal(e.total);
      } catch {
        // leave empty
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [address]);

  const handleUpdate = async (updates: { Status?: string; Category?: string }) => {
    if (!detail) return;
    setUpdating(true);
    try {
      await apiPatch("senders", {
        FromAddress: address,
        Status: updates.Status ?? detail.Status,
        Category: updates.Category ?? (detail.Category !== detail.AutoCategory ? detail.Category : ""),
      });
      setDetail((prev) =>
        prev
          ? {
              ...prev,
              ...(updates.Status !== undefined && { Status: updates.Status }),
              ...(updates.Category !== undefined && {
                Category: updates.Category || prev.AutoCategory,
              }),
            }
          : prev
      );
    } catch {
      // keep state
    } finally {
      setUpdating(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center gap-2 py-12 justify-center text-muted-foreground">
        <Loader2 className="size-5 animate-spin" />
        Loading sender details...
      </div>
    );
  }

  if (!detail) {
    return (
      <div className="space-y-4">
        <nav className="flex items-center gap-1.5 text-sm text-muted-foreground">
          <Link href="/senders" className="hover:text-foreground transition-colors">Senders</Link>
          <span>/</span>
          <span className="text-foreground">Not found</span>
        </nav>
        <p className="text-muted-foreground">Sender not found.</p>
      </div>
    );
  }

  const name = detail.FromName || detail.FromAddress;
  const isOverridden = detail.Category !== detail.AutoCategory && detail.AutoCategory !== "";

  // Calculate average frequency from history data
  const weeksWithData = history.filter((w) => w.Count > 0).length;
  const totalFromHistory = history.reduce((sum, w) => sum + w.Count, 0);
  const avgPerWeek = weeksWithData > 0 ? totalFromHistory / history.length : 0;

  // Days since last email
  const daysSinceLast = detail.LastSeen
    ? Math.floor((Date.now() - new Date(detail.LastSeen).getTime()) / 86400000)
    : null;

  return (
    <div className="space-y-6">
      <nav className="flex items-center gap-1.5 text-sm text-muted-foreground">
        <Link href="/senders" className="hover:text-foreground transition-colors">Senders</Link>
        <span>/</span>
        <span className="text-foreground truncate max-w-xs">{name}</span>
      </nav>

      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold">{name}</h1>
          {detail.FromName && (
            <p className="text-sm text-muted-foreground">{detail.FromAddress}</p>
          )}
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted-foreground">
            <span className="flex items-center gap-1">
              <Mail className="size-3.5" />
              {detail.EmailCount} email{detail.EmailCount !== 1 ? "s" : ""}
            </span>
            <span className="flex items-center gap-1">
              <Calendar className="size-3.5" />
              First seen {formatShortDate(detail.FirstSeen)}
            </span>
            <span className="flex items-center gap-1">
              <Clock className="size-3.5" />
              Last seen{" "}
              {daysSinceLast === 0
                ? "today"
                : daysSinceLast === 1
                  ? "yesterday"
                  : `${formatShortDate(detail.LastSeen)}`}
            </span>
            {avgPerWeek > 0 && (
              <span className="flex items-center gap-1">
                <TrendingUp className="size-3.5" />
                ~{avgPerWeek < 1 ? avgPerWeek.toFixed(1) : Math.round(avgPerWeek)}/week
              </span>
            )}
          </div>
          <div className="flex items-center gap-2 pt-1">
            {detail.Category && (
              <span
                className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
                  CATEGORY_COLORS[detail.Category] || CATEGORY_COLORS.other
                }`}
              >
                {detail.Category}
                {isOverridden && " (override)"}
              </span>
            )}
            <Link
              href={`/inbox?search=${encodeURIComponent(address)}`}
              className="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary hover:bg-primary/20 transition-colors"
            >
              <ExternalLink className="size-3" />
              View in inbox
            </Link>
          </div>
        </div>
      </div>

      {/* Controls */}
      <div className="flex items-center gap-3">
        <div className="space-y-1">
          <label className="text-xs font-medium text-muted-foreground">Status</label>
          <select
            value={detail.Status}
            onChange={(e) => handleUpdate({ Status: e.target.value })}
            disabled={updating}
            className="block h-9 rounded-md border border-input bg-transparent px-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-1">
          <label className="text-xs font-medium text-muted-foreground">Category</label>
          <select
            value={detail.Category === detail.AutoCategory ? "" : detail.Category}
            onChange={(e) => handleUpdate({ Category: e.target.value })}
            disabled={updating}
            className="block h-9 rounded-md border border-input bg-transparent px-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          >
            {CATEGORY_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.value === "" && detail.AutoCategory
                  ? `Auto (${detail.AutoCategory})`
                  : opt.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Volume chart */}
      {history.length > 0 && (
        <div className="rounded-lg border p-4 space-y-3">
          <h2 className="flex items-center gap-2 text-sm font-medium">
            <BarChart3 className="size-4" />
            Email Volume (12 weeks)
          </h2>
          <VolumeChart data={history} />
        </div>
      )}

      {/* Recent emails */}
      <div className="space-y-3">
        <h2 className="text-sm font-medium">
          Recent Emails {emailTotal > 0 && <span className="text-muted-foreground">({emailTotal} total)</span>}
        </h2>
        {emails.length === 0 ? (
          <p className="text-sm text-muted-foreground">No emails from this sender.</p>
        ) : (
          <div className="divide-y rounded-lg border">
            {emails.map((email) => (
              <Link
                key={email.ID}
                href={`/inbox/${email.ID}`}
                className="flex items-center justify-between gap-3 p-3 hover:bg-accent/50 transition-colors"
              >
                <div className="min-w-0 flex-1">
                  <p className={`text-sm truncate ${email.IsRead ? "text-muted-foreground" : "font-medium"}`}>
                    {email.Subject || "(no subject)"}
                  </p>
                </div>
                <span className="text-xs text-muted-foreground whitespace-nowrap shrink-0">
                  {formatShortDate(email.ReceivedAt)}
                </span>
              </Link>
            ))}
          </div>
        )}
        {emailTotal > 10 && (
          <Link
            href={`/inbox?search=${encodeURIComponent(address)}`}
            className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
          >
            View all {emailTotal} emails in inbox
            <ExternalLink className="size-3" />
          </Link>
        )}
      </div>
    </div>
  );
}
