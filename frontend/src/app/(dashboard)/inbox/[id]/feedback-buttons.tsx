"use client";

import { useState, useEffect } from "react";
import { apiGet, apiPost, apiDelete } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { ThumbsUp, ThumbsDown } from "lucide-react";
import { cn } from "@/lib/utils";
import type { EmailFeedback } from "@/types/api";

interface FeedbackButtonsProps {
  emailId: string;
}

export function FeedbackButtons({ emailId }: FeedbackButtonsProps) {
  const [rating, setRating] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    apiGet<EmailFeedback>(`emails/${emailId}/feedback`)
      .then((fb) => {
        if (fb.Rating) setRating(fb.Rating);
      })
      .catch(() => {});
  }, [emailId]);

  const handleFeedback = async (value: "useful" | "not_useful") => {
    setLoading(true);
    const prevRating = rating;
    const newRating = rating === value ? null : value;
    setRating(newRating);
    try {
      if (newRating === null) {
        await apiDelete(`emails/${emailId}/feedback`);
      } else {
        await apiPost(`emails/${emailId}/feedback`, { Rating: newRating });
      }
    } catch {
      setRating(prevRating);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <span className="text-sm text-muted-foreground">Was this useful?</span>
      <Button
        variant="ghost"
        size="sm"
        disabled={loading}
        onClick={() => handleFeedback("useful")}
        className={cn(
          "h-8 gap-1",
          rating === "useful" && "bg-success/10 text-success hover:bg-success/20"
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
          rating === "not_useful" && "bg-destructive/10 text-destructive hover:bg-destructive/20"
        )}
      >
        <ThumbsDown className="size-4" />
        No
      </Button>
    </div>
  );
}
