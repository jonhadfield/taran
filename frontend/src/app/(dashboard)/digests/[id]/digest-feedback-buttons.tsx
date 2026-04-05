"use client";

import { useState, useEffect } from "react";
import { apiGet, apiPost, apiDelete } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { ThumbsUp, ThumbsDown } from "lucide-react";
import { cn } from "@/lib/utils";
import type { DigestFeedback } from "@/types/api";

interface DigestFeedbackButtonsProps {
  digestId: string;
}

export function DigestFeedbackButtons({
  digestId,
}: DigestFeedbackButtonsProps) {
  const [rating, setRating] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    apiGet<DigestFeedback>(`digests/${digestId}/feedback`)
      .then((fb) => {
        if (fb.Rating) setRating(fb.Rating);
      })
      .catch(() => {});
  }, [digestId]);

  const handleFeedback = async (value: "useful" | "not_useful") => {
    setLoading(true);
    const prevRating = rating;
    const newRating = rating === value ? null : value;
    setRating(newRating);
    try {
      if (newRating === null) {
        await apiDelete(`digests/${digestId}/feedback`);
      } else {
        await apiPost(`digests/${digestId}/feedback`, { Rating: newRating });
      }
    } catch {
      setRating(prevRating);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <span className="text-sm text-muted-foreground">
        Was this digest useful?
      </span>
      <Button
        variant="ghost"
        size="sm"
        disabled={loading}
        onClick={() => handleFeedback("useful")}
        className={cn(
          "h-8 gap-1",
          rating === "useful" &&
            "bg-green-500/10 text-green-600 hover:bg-green-500/20"
        )}
      >
        <ThumbsUp className="size-4" />
        Yes
      </Button>
      <Button
        variant="ghost"
        size="sm"
        disabled={loading}
        onClick={() => handleFeedback("not_useful")}
        className={cn(
          "h-8 gap-1",
          rating === "not_useful" &&
            "bg-red-500/10 text-red-600 hover:bg-red-500/20"
        )}
      >
        <ThumbsDown className="size-4" />
        No
      </Button>
    </div>
  );
}
