"use client";

import { useState } from "react";
import { apiPost } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Mail, Check } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";

interface SendEmailButtonProps {
  digestId: string;
  alreadySent: boolean;
}

export function SendEmailButton({ digestId, alreadySent }: SendEmailButtonProps) {
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(alreadySent);
  const [error, setError] = useState<string | null>(null);

  const handleSend = async () => {
    setLoading(true);
    setError(null);
    try {
      await apiPost(`admin/digests/${digestId}/send`, {});
      setSent(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to send email");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <Button
        variant={sent ? "outline" : "default"}
        size="sm"
        onClick={handleSend}
        disabled={loading}
      >
        {loading ? (
          <Spinner className="mr-1.5" />
        ) : sent ? (
          <Check className="size-4 mr-1.5" />
        ) : (
          <Mail className="size-4 mr-1.5" />
        )}
        {sent ? "Resend Email" : "Send Email"}
      </Button>
      {error && <span className="text-sm text-destructive">{error}</span>}
    </div>
  );
}
