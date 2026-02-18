"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { RefreshCw } from "lucide-react";
import { apiPost } from "@/lib/api";

export function ReprocessButton({ emailId }: { emailId: string }) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleClick() {
    setLoading(true);
    setError(null);
    try {
      await apiPost(`emails/${emailId}/reprocess`, {});
      router.refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to reprocess");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mt-2">
      <Button
        variant="outline"
        size="sm"
        onClick={handleClick}
        disabled={loading}
      >
        <RefreshCw className={`size-4 mr-1.5 ${loading ? "animate-spin" : ""}`} />
        {loading ? "Reprocessing..." : "Reprocess"}
      </Button>
      {error && <p className="text-sm text-destructive mt-1">{error}</p>}
    </div>
  );
}
