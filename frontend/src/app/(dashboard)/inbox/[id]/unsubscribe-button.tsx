"use client";

import { useState } from "react";
import { toast } from "sonner";
import { apiPost } from "@/lib/api";
import { isSafeURL } from "@/lib/utils";
import { MailX } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";

interface UnsubscribeButtonProps {
  emailId: string;
}

export function UnsubscribeButton({ emailId }: UnsubscribeButtonProps) {
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);

  const handleUnsubscribe = async () => {
    setLoading(true);
    try {
      const res = await apiPost<{ status: string; url?: string; address?: string; method?: string }>(
        `emails/${emailId}/unsubscribe`, {}
      );

      if (res.status === "unsubscribed") {
        toast.success("Successfully unsubscribed");
        setDone(true);
      } else if (res.status === "redirect" && res.url) {
        if (isSafeURL(res.url)) {
          window.open(res.url, "_blank", "noopener,noreferrer");
          toast.info("Unsubscribe page opened in a new tab");
        } else {
          toast.error("Invalid unsubscribe URL");
        }
        setDone(true);
      } else if (res.status === "mailto" && res.address) {
        window.location.href = `mailto:${res.address}?subject=unsubscribe`;
        toast.info("Opening email client to unsubscribe");
        setDone(true);
      }
    } catch {
      toast.error("Failed to unsubscribe");
    } finally {
      setLoading(false);
    }
  };

  return (
    <button
      onClick={handleUnsubscribe}
      disabled={loading || done}
      className="inline-flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs font-medium transition-colors hover:bg-accent disabled:opacity-50"
    >
      {loading ? (
        <Spinner className="size-3.5" />
      ) : (
        <MailX className="size-3.5" />
      )}
      {done ? "Unsubscribed" : "Unsubscribe"}
    </button>
  );
}
