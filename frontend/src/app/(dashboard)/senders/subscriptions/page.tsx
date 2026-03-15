"use client";

import { useState, useEffect, useCallback } from "react";
import { apiGet, apiPost } from "@/lib/api";
import type { SubscriptionInfo } from "@/types/api";
import { ArrowLeft, MailX, Check, Loader2, ExternalLink } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import Link from "next/link";

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

export default function SubscriptionsPage() {
  const [subs, setSubs] = useState<SubscriptionInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedAddresses, setSelectedAddresses] = useState<Set<string>>(new Set());
  const [unsubscribing, setUnsubscribing] = useState<string | null>(null);
  const [batchLoading, setBatchLoading] = useState(false);

  const fetchSubs = useCallback(async () => {
    try {
      const data = await apiGet<SubscriptionInfo[]>("subscriptions");
      setSubs(data || []);
    } catch {
      toast.error("Failed to load subscriptions");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSubs();
  }, [fetchSubs]);

  const handleUnsubscribe = async (fromAddress: string) => {
    setUnsubscribing(fromAddress);
    try {
      const results = await apiPost<{ address: string; status: string; url?: string }[]>(
        "subscriptions/unsubscribe",
        { FromAddresses: [fromAddress] }
      );
      const result = results?.[0];
      if (result?.status === "redirect" && result.url) {
        window.open(result.url, "_blank", "noopener,noreferrer");
        toast.success("Opened unsubscribe page in new tab");
      } else if (result?.status === "unsubscribed") {
        toast.success("Unsubscribed successfully");
      } else if (result?.status === "mailto") {
        toast.info("Unsubscribe requires sending an email — marked as unsubscribed");
      } else {
        toast.success("Unsubscribe initiated");
      }
      setSubs((prev) =>
        prev.map((s) =>
          s.FromAddress === fromAddress
            ? { ...s, UnsubscribedAt: new Date().toISOString() }
            : s
        )
      );
    } catch {
      toast.error("Failed to unsubscribe");
    } finally {
      setUnsubscribing(null);
    }
  };

  const handleBatchUnsubscribe = async () => {
    if (selectedAddresses.size === 0) return;
    setBatchLoading(true);
    try {
      const results = await apiPost<{ address: string; status: string; url?: string }[]>(
        "subscriptions/unsubscribe",
        { FromAddresses: Array.from(selectedAddresses) }
      );
      const redirects = results?.filter((r) => r.status === "redirect") || [];
      if (redirects.length > 0 && redirects.length <= 3) {
        for (const r of redirects) {
          if (r.url) window.open(r.url, "_blank", "noopener,noreferrer");
        }
      }
      toast.success(`Unsubscribed from ${selectedAddresses.size} sender${selectedAddresses.size > 1 ? "s" : ""}`);
      setSubs((prev) =>
        prev.map((s) =>
          selectedAddresses.has(s.FromAddress)
            ? { ...s, UnsubscribedAt: new Date().toISOString() }
            : s
        )
      );
      setSelectedAddresses(new Set());
    } catch {
      toast.error("Failed to batch unsubscribe");
    } finally {
      setBatchLoading(false);
    }
  };

  const toggleSelect = (addr: string) => {
    setSelectedAddresses((prev) => {
      const next = new Set(prev);
      if (next.has(addr)) next.delete(addr);
      else next.add(addr);
      return next;
    });
  };

  const activeSubs = subs.filter((s) => !s.UnsubscribedAt);
  const unsubscribedSubs = subs.filter((s) => s.UnsubscribedAt);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          href="/senders"
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground py-1"
        >
          <ArrowLeft className="size-4" />
          Back to Senders
        </Link>
      </div>

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Manage Subscriptions</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Senders with unsubscribe links from your emails
          </p>
        </div>
        {selectedAddresses.size > 0 && (
          <Button
            onClick={handleBatchUnsubscribe}
            disabled={batchLoading}
            variant="destructive"
            size="sm"
          >
            {batchLoading ? (
              <Loader2 className="size-4 animate-spin mr-1.5" />
            ) : (
              <MailX className="size-4 mr-1.5" />
            )}
            Unsubscribe ({selectedAddresses.size})
          </Button>
        )}
      </div>

      {subs.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <MailX className="size-12 text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium">No subscriptions found</h3>
          <p className="text-sm text-muted-foreground mt-2">
            We&apos;ll show senders here once we detect unsubscribe links in your emails.
          </p>
        </div>
      ) : (
        <>
          {activeSubs.length > 0 && (
            <div className="space-y-2">
              <h2 className="text-sm font-medium text-muted-foreground">
                Active ({activeSubs.length})
              </h2>
              <div className="divide-y rounded-lg border">
                {activeSubs.map((sub) => (
                  <SubscriptionRow
                    key={sub.FromAddress}
                    sub={sub}
                    isSelected={selectedAddresses.has(sub.FromAddress)}
                    onToggleSelect={() => toggleSelect(sub.FromAddress)}
                    onUnsubscribe={() => handleUnsubscribe(sub.FromAddress)}
                    isUnsubscribing={unsubscribing === sub.FromAddress}
                  />
                ))}
              </div>
            </div>
          )}

          {unsubscribedSubs.length > 0 && (
            <div className="space-y-2">
              <h2 className="text-sm font-medium text-muted-foreground">
                Unsubscribed ({unsubscribedSubs.length})
              </h2>
              <div className="divide-y rounded-lg border opacity-60">
                {unsubscribedSubs.map((sub) => (
                  <SubscriptionRow
                    key={sub.FromAddress}
                    sub={sub}
                    isSelected={false}
                    onToggleSelect={() => {}}
                    onUnsubscribe={() => {}}
                    isUnsubscribing={false}
                  />
                ))}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}

function SubscriptionRow({
  sub,
  isSelected,
  onToggleSelect,
  onUnsubscribe,
  isUnsubscribing,
}: {
  sub: SubscriptionInfo;
  isSelected: boolean;
  onToggleSelect: () => void;
  onUnsubscribe: () => void;
  isUnsubscribing: boolean;
}) {
  const senderName = sub.FromName || sub.FromAddress;
  const initial = senderName.charAt(0).toUpperCase();
  const isUnsubscribed = !!sub.UnsubscribedAt;

  return (
    <div className={`flex items-center gap-3 p-3.5 transition-colors hover:bg-accent/50 ${isSelected ? "bg-accent/30" : ""}`}>
      {!isUnsubscribed && (
        <input
          type="checkbox"
          className="size-4 shrink-0 rounded border-input accent-primary cursor-pointer"
          checked={isSelected}
          onChange={onToggleSelect}
          aria-label={`Select ${senderName}`}
        />
      )}

      <div
        className={`hidden sm:flex size-9 shrink-0 items-center justify-center rounded-full text-sm font-medium text-white ${getAvatarColor(senderName)}`}
      >
        {initial}
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium truncate">{senderName}</span>
          {isUnsubscribed && (
            <Badge variant="outline" className="text-xs shrink-0">
              <Check className="size-3 mr-1" />
              Unsubscribed
            </Badge>
          )}
        </div>
        <p className="text-xs text-muted-foreground truncate">{sub.FromAddress}</p>
        <p className="text-xs text-muted-foreground">
          {sub.EmailCount} email{sub.EmailCount !== 1 ? "s" : ""}
          {" \u00b7 "}
          Last: {new Date(sub.LastSeen).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}
        </p>
      </div>

      <div className="flex items-center gap-2 shrink-0">
        {sub.UnsubscribeURL && !isUnsubscribed && (
          <a
            href={sub.UnsubscribeURL}
            target="_blank"
            rel="noopener noreferrer"
            className="text-muted-foreground hover:text-foreground transition-colors"
            title="Open unsubscribe page"
          >
            <ExternalLink className="size-4" />
          </a>
        )}
        {!isUnsubscribed && (
          <Button
            variant="outline"
            size="sm"
            onClick={onUnsubscribe}
            disabled={isUnsubscribing}
            className="text-xs"
          >
            {isUnsubscribing ? (
              <Loader2 className="size-3.5 animate-spin" />
            ) : (
              "Unsubscribe"
            )}
          </Button>
        )}
      </div>
    </div>
  );
}
