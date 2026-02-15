"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { apiPost } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Loader2, Sparkles } from "lucide-react";
import type { Digest } from "@/types/api";

export function GenerateDigestButton() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleGenerate = async () => {
    setLoading(true);
    setError(null);
    try {
      const digest = await apiPost<Digest>("digests/generate", {});
      router.push(`/digests/${digest.ID}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to generate digest");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <Button onClick={handleGenerate} disabled={loading} size="sm">
        {loading ? (
          <Loader2 className="size-4 animate-spin mr-1.5" />
        ) : (
          <Sparkles className="size-4 mr-1.5" />
        )}
        Generate Digest
      </Button>
      {error && <span className="text-sm text-destructive">{error}</span>}
    </div>
  );
}
