"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { apiPost } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { BarChart3 } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import type { WeeklySummary } from "@/types/api";

export function GenerateWeeklyButton() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleGenerate = async () => {
    setLoading(true);
    setError(null);
    try {
      const summary = await apiPost<WeeklySummary>("weekly-summaries/generate", {});
      router.push(`/digests/weekly/${summary.ID}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to generate summary");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <Button onClick={handleGenerate} disabled={loading} size="sm" variant="outline">
        {loading ? (
          <Spinner className="mr-1.5" />
        ) : (
          <BarChart3 className="size-4 mr-1.5" />
        )}
        Generate Weekly Summary
      </Button>
      {error && <span className="text-sm text-destructive">{error}</span>}
    </div>
  );
}
