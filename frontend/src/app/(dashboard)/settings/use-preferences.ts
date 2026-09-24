"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { apiGet, apiPatch } from "@/lib/api";
import { useColorTheme } from "@/components/color-theme-provider";
import type { UserPreference } from "@/types/api";

const COMMON_TIMEZONES = [
  "America/New_York",
  "America/Chicago",
  "America/Denver",
  "America/Los_Angeles",
  "America/Anchorage",
  "Pacific/Honolulu",
  "Europe/London",
  "Europe/Paris",
  "Europe/Berlin",
  "Asia/Tokyo",
  "Asia/Shanghai",
  "Asia/Kolkata",
  "Australia/Sydney",
  "Pacific/Auckland",
  "UTC",
];

function detectTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone;
  } catch {
    return "UTC";
  }
}

export const DEFAULT_PREF: UserPreference = {
  UserID: "", DigestEmail: false, DigestFrequency: "daily", DigestHour: 7,
  DigestDay: 1, DigestTimezone: "UTC", TopicLimit: 15, DigestStyle: "detailed",
  InterestKeywords: [], ExclusionKeywords: [], ColorTheme: "neutral",
  ExcludedCategories: ["notification", "transactional", "marketing"],
  DailyTokenLimit: 0, QuietHoursEnabled: false, QuietHoursStart: 22,
  QuietHoursEnd: 7, WeeklySummary: true, DigestWebhook: false, WebhookURL: "",
  CreatedAt: "", UpdatedAt: "",
};

/**
 * Loads the user's preferences and saves changes as they are made.
 *
 * Each settings page loads them itself, so a page is self-contained and can be
 * linked to directly.
 */
export function usePreferences() {
  const { setColorTheme } = useColorTheme();
  const [pref, setPref] = useState<UserPreference>(DEFAULT_PREF);
  const [prefLoading, setPrefLoading] = useState(true);
  const [prefSaving, setPrefSaving] = useState(false);

  const fetchPreferences = useCallback(async () => {
    try {
      const data = await apiGet<UserPreference>("preferences");
      setPref({
        ...DEFAULT_PREF,
        ...data,
        DigestTimezone: data.DigestTimezone || detectTimezone(),
        InterestKeywords: data.InterestKeywords || [],
        ExclusionKeywords: data.ExclusionKeywords || [],
        ExcludedCategories: data.ExcludedCategories || ["notification", "transactional", "marketing"],
      });
      if (data.ColorTheme) {
        setColorTheme(data.ColorTheme as Parameters<typeof setColorTheme>[0]);
      }
    } catch {
      setPref((p) => ({ ...p, DigestTimezone: detectTimezone() }));
    } finally {
      setPrefLoading(false);
    }
  }, [setColorTheme]);

  useEffect(() => {
    fetchPreferences();
  }, [fetchPreferences]);

  const updatePreference = useCallback(
    async (updates: Partial<UserPreference>) => {
      setPrefSaving(true);
      try {
        const updated = await apiPatch<UserPreference>("preferences", updates);
        setPref((prev) => ({ ...prev, ...updated }));
        toast.success("Preferences saved");
      } catch {
        toast.error("Failed to save preferences");
        await fetchPreferences();
      } finally {
        setPrefSaving(false);
      }
    },
    [fetchPreferences],
  );

  // Show the change straight away, then persist it.
  const handlePrefChange = useCallback(
    (updates: Partial<UserPreference>) => {
      setPref((prev) => ({ ...prev, ...updates }));
      updatePreference(updates);
    },
    [updatePreference],
  );

  const timezoneOptions = useMemo(() => {
    const tz = detectTimezone();
    return COMMON_TIMEZONES.includes(tz) ? COMMON_TIMEZONES : [tz, ...COMMON_TIMEZONES];
  }, []);

  return { pref, setPref, prefLoading, prefSaving, updatePreference, handlePrefChange, timezoneOptions };
}
