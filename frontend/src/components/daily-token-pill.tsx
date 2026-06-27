"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { apiGet } from "@/lib/api";
import type { UsageStats } from "@/types/api";
import { cn, formatTokens, tokenTextColor } from "@/lib/utils";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

/**
 * Compact header indicator showing how many AI tokens remain against the
 * user's daily limit. Renders nothing unless a daily limit is configured.
 */
export function DailyTokenPill() {
  const [stats, setStats] = useState<UsageStats | null>(null);

  const load = useCallback(() => {
    apiGet<UsageStats>("usage")
      .then(setStats)
      .catch(() => {});
  }, []);

  useEffect(() => {
    load();
    const onFocus = () => load();
    window.addEventListener("focus", onFocus);
    return () => window.removeEventListener("focus", onFocus);
  }, [load]);

  // No daily limit set → nothing to show.
  if (!stats || stats.DailyTokenLimit <= 0) return null;

  const used = stats.DailyTokensUsed;
  const limit = stats.DailyTokenLimit;
  const remaining = Math.max(0, limit - used);
  const percent = Math.min(100, (used / limit) * 100);
  const atLimit = remaining === 0;

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Link
            href="/settings"
            aria-label={`${formatTokens(remaining)} of ${formatTokens(limit)} daily tokens remaining`}
            className={cn(
              "hidden sm:inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium transition-colors hover:bg-muted",
              tokenTextColor(percent)
            )}
          >
            <span
              className={cn(
                "size-1.5 rounded-full",
                percent >= 90 ? "bg-red-500" : percent >= 70 ? "bg-yellow-500" : "bg-primary"
              )}
            />
            {atLimit ? "Daily limit reached" : `${formatTokens(remaining)} left today`}
          </Link>
        </TooltipTrigger>
        <TooltipContent>
          {formatTokens(used)} / {formatTokens(limit)} tokens used today
          {" · "}
          {Math.round(percent)}%
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
