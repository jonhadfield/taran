"use client";

import { useState } from "react";
import { apiPost, apiDelete } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Check, Copy, Link2, Link2Off, Loader2 } from "lucide-react";

interface ShareButtonProps {
  digestId: string;
  initialToken: string | null;
}

export function ShareButton({ digestId, initialToken }: ShareButtonProps) {
  const [token, setToken] = useState<string | null>(initialToken);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  const shareUrl = token ? `${window.location.origin}/shared/${token}` : null;

  const handleShare = async () => {
    setLoading(true);
    try {
      const res = await apiPost<{ ShareToken: string }>(`digests/${digestId}/share`, {});
      setToken(res.ShareToken);
    } catch {
      // Ignore
    } finally {
      setLoading(false);
    }
  };

  const handleUnshare = async () => {
    setLoading(true);
    try {
      await apiDelete(`digests/${digestId}/share`);
      setToken(null);
    } catch {
      // Ignore
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    if (!shareUrl) return;
    await navigator.clipboard.writeText(shareUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (!token) {
    return (
      <Button variant="outline" size="sm" onClick={handleShare} disabled={loading}>
        {loading ? <Loader2 className="mr-1 h-4 w-4 animate-spin" /> : <Link2 className="mr-1 h-4 w-4" />}
        Share
      </Button>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="sm" onClick={handleCopy}>
        {copied ? <Check className="mr-1 h-4 w-4" /> : <Copy className="mr-1 h-4 w-4" />}
        {copied ? "Copied" : "Copy link"}
      </Button>
      <Button variant="ghost" size="sm" onClick={handleUnshare} disabled={loading}>
        {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Link2Off className="h-4 w-4" />}
      </Button>
    </div>
  );
}
