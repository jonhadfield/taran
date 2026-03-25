"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { apiGet, apiDelete, apiPatch } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { EmailAccount, ListResponse, UserPreference } from "@/types/api";
import { SignOutButton } from "./sign-out-button";
import { ExportDataButton } from "./export-data-button";
import { ForwardingGuide } from "@/components/forwarding-guide";
import { AccountSettings } from "./account-settings";
import { ThemeColorSettings } from "./theme-color-settings";
import { DigestDeliverySettings } from "./digest-delivery-settings";
import { InboxDisplaySettings } from "./inbox-display-settings";
import { DigestStyleSettings } from "./digest-style-settings";
import { KeywordPreferencesSettings } from "./keyword-preferences-settings";
import { DigestCategoriesSettings } from "./digest-categories-settings";
import { UsageStatsCard } from "./usage-stats";
import { ApiKeysSettings } from "./api-keys-settings";
import { QuietHoursSettings } from "./quiet-hours-settings";
import { AutoArchiveSettings } from "./auto-archive-settings";
import { LabelSettings } from "./label-settings";
import { DailyLimitSettings } from "./daily-limit-settings";
import { useColorTheme } from "@/components/color-theme-provider";
import { SettingsNav, type SettingsSection } from "./settings-nav";

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

const SECTIONS: SettingsSection[] = [
  { id: "accounts", label: "Accounts", group: "General" },
  { id: "appearance", label: "Appearance", group: "General" },
  { id: "delivery", label: "Delivery", group: "Digest" },
  { id: "quiet-hours", label: "Quiet Hours", group: "Digest" },
  { id: "digest-style", label: "Digest Style", group: "Digest" },
  { id: "categories", label: "Categories", group: "Digest" },
  { id: "keywords", label: "Keywords", group: "Digest" },
  { id: "inbox", label: "Inbox", group: "Organisation" },
  { id: "labels", label: "Labels", group: "Organisation" },
  { id: "auto-archive", label: "Auto-Archive", group: "Organisation" },
  { id: "forwarding", label: "Forwarding", group: "Advanced" },
  { id: "api-keys", label: "API Keys", group: "Advanced" },
  { id: "limits", label: "Limits", group: "Advanced" },
  { id: "usage", label: "Usage", group: "Advanced" },
  { id: "data", label: "Data", group: "Advanced" },
  { id: "account", label: "Account", group: "Advanced" },
];

export default function SettingsPage() {
  const { colorTheme, setColorTheme } = useColorTheme();
  const [accounts, setAccounts] = useState<EmailAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Preferences
  const [digestEmail, setDigestEmail] = useState(false);
  const [digestFrequency, setDigestFrequency] = useState("daily");
  const [digestHour, setDigestHour] = useState(7);
  const [digestDay, setDigestDay] = useState(1);
  const [digestTimezone, setDigestTimezone] = useState("UTC");
  const [topicLimit, setTopicLimit] = useState(15);
  const [digestStyle, setDigestStyle] = useState("detailed");
  const [interestKeywords, setInterestKeywords] = useState<string[]>([]);
  const [exclusionKeywords, setExclusionKeywords] = useState<string[]>([]);
  const [excludedCategories, setExcludedCategories] = useState<string[]>(["notification", "transactional", "marketing"]);
  const [dailyTokenLimit, setDailyTokenLimit] = useState(0);
  const [quietHoursEnabled, setQuietHoursEnabled] = useState(false);
  const [quietHoursStart, setQuietHoursStart] = useState(22);
  const [quietHoursEnd, setQuietHoursEnd] = useState(7);
  const [digestWebhook, setDigestWebhook] = useState(false);
  const [webhookURL, setWebhookURL] = useState("");
  const [weeklySummary, setWeeklySummary] = useState(true);
  const [prefLoading, setPrefLoading] = useState(true);
  const [prefSaving, setPrefSaving] = useState(false);

  // Delete confirmation
  const [deleteTarget, setDeleteTarget] = useState<EmailAccount | null>(null);
  const [deleting, setDeleting] = useState(false);

  const fetchAccounts = useCallback(async () => {
    try {
      const res = await apiGet<ListResponse<EmailAccount>>("accounts");
      setAccounts(res.data || []);
      setError("");
    } catch {
      setError("Failed to load accounts");
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchPreferences = useCallback(async () => {
    try {
      const pref = await apiGet<UserPreference>("preferences");
      setDigestEmail(pref.DigestEmail);
      setDigestFrequency(pref.DigestFrequency || "daily");
      setDigestHour(pref.DigestHour ?? 7);
      setDigestDay(pref.DigestDay ?? 1);
      setDigestTimezone(pref.DigestTimezone || detectTimezone());
      setTopicLimit(pref.TopicLimit ?? 15);
      setDigestStyle(pref.DigestStyle || "detailed");
      setInterestKeywords(pref.InterestKeywords || []);
      setExclusionKeywords(pref.ExclusionKeywords || []);
      setExcludedCategories(pref.ExcludedCategories || ["notification", "transactional", "marketing"]);
      setDailyTokenLimit(pref.DailyTokenLimit ?? 0);
      setQuietHoursEnabled(pref.QuietHoursEnabled ?? false);
      setQuietHoursStart(pref.QuietHoursStart ?? 22);
      setQuietHoursEnd(pref.QuietHoursEnd ?? 7);
      setDigestWebhook(pref.DigestWebhook ?? false);
      setWebhookURL(pref.WebhookURL ?? "");
      setWeeklySummary(pref.WeeklySummary ?? true);
      if (pref.ColorTheme) {
        setColorTheme(pref.ColorTheme as Parameters<typeof setColorTheme>[0]);
      }
    } catch {
      setDigestTimezone(detectTimezone());
    } finally {
      setPrefLoading(false);
    }
  }, [setColorTheme]);

  useEffect(() => {
    fetchAccounts();
    fetchPreferences();
  }, [fetchAccounts, fetchPreferences]);

  const updatePreference = async (updates: Partial<Pick<UserPreference, "DigestEmail" | "DigestFrequency" | "DigestHour" | "DigestDay" | "DigestTimezone" | "TopicLimit" | "DigestStyle" | "InterestKeywords" | "ExclusionKeywords" | "ColorTheme" | "ExcludedCategories" | "DailyTokenLimit" | "QuietHoursEnabled" | "QuietHoursStart" | "QuietHoursEnd" | "DigestWebhook" | "WebhookURL" | "WeeklySummary">>) => {
    setPrefSaving(true);
    try {
      const updated = await apiPatch<UserPreference>("preferences", updates);
      setDigestEmail(updated.DigestEmail);
      setDigestFrequency(updated.DigestFrequency || "daily");
      setDigestHour(updated.DigestHour ?? 7);
      setDigestDay(updated.DigestDay ?? 1);
      setDigestTimezone(updated.DigestTimezone || "UTC");
      setTopicLimit(updated.TopicLimit ?? 15);
      setDigestStyle(updated.DigestStyle || "detailed");
      setInterestKeywords(updated.InterestKeywords || []);
      setExclusionKeywords(updated.ExclusionKeywords || []);
      setExcludedCategories(updated.ExcludedCategories || ["notification", "transactional", "marketing"]);
      setDailyTokenLimit(updated.DailyTokenLimit ?? 0);
      setQuietHoursEnabled(updated.QuietHoursEnabled ?? false);
      setQuietHoursStart(updated.QuietHoursStart ?? 22);
      setQuietHoursEnd(updated.QuietHoursEnd ?? 7);
      setDigestWebhook(updated.DigestWebhook ?? false);
      setWebhookURL(updated.WebhookURL ?? "");
      setWeeklySummary(updated.WeeklySummary ?? true);
      toast.success("Preferences saved");
    } catch {
      toast.error("Failed to save preferences");
      await fetchPreferences();
    } finally {
      setPrefSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);

    try {
      await apiDelete(`accounts/${deleteTarget.ID}`);
      setDeleteTarget(null);
      toast.success("Inbox deleted");
      await fetchAccounts();
    } catch {
      toast.error("Failed to delete inbox");
      setDeleteTarget(null);
    } finally {
      setDeleting(false);
    }
  };

  // Build timezone list: user's detected TZ first, then common ones
  const browserTz = detectTimezone();
  const timezoneOptions = COMMON_TIMEZONES.includes(browserTz)
    ? COMMON_TIMEZONES
    : [browserTz, ...COMMON_TIMEZONES];

  // Filter sections: hide "Forwarding" if no accounts yet
  const visibleSections = accounts.length > 0
    ? SECTIONS
    : SECTIONS.filter((s) => s.id !== "forwarding");

  return (
    <div className="max-w-2xl">
      <h1 className="text-2xl font-bold">Settings</h1>

      <SettingsNav sections={visibleSections} />

      <div className="space-y-6 mt-6">

        <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground pt-2">General</h2>

        <section id="accounts" className="scroll-mt-24">
          <AccountSettings
            accounts={accounts}
            loading={loading}
            error={error}
            deleteTarget={deleteTarget}
            deleting={deleting}
            onFetchAccounts={fetchAccounts}
            onSetDeleteTarget={setDeleteTarget}
            onDelete={handleDelete}
          />
        </section>

        <section id="appearance" className="scroll-mt-24">
          <ThemeColorSettings
            colorTheme={colorTheme}
            onColorThemeChange={setColorTheme}
          />
        </section>

        <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground pt-4">Digest</h2>

        <section id="delivery" className="scroll-mt-24">
          <DigestDeliverySettings
            digestEmail={digestEmail}
            digestFrequency={digestFrequency}
            digestHour={digestHour}
            digestDay={digestDay}
            digestTimezone={digestTimezone}
            timezoneOptions={timezoneOptions}
            prefLoading={prefLoading}
            prefSaving={prefSaving}
            onToggleDigestEmail={(checked) => {
              setDigestEmail(checked);
              updatePreference({ DigestEmail: checked });
            }}
            onFrequencyChange={(value) => {
              setDigestFrequency(value);
              updatePreference({ DigestFrequency: value });
            }}
            onHourChange={(value) => {
              setDigestHour(value);
              updatePreference({ DigestHour: value });
            }}
            onDayChange={(value) => {
              setDigestDay(value);
              updatePreference({ DigestDay: value });
            }}
            onTimezoneChange={(value) => {
              setDigestTimezone(value);
              updatePreference({ DigestTimezone: value });
            }}
            digestWebhook={digestWebhook}
            webhookURL={webhookURL}
            onToggleDigestWebhook={(checked) => {
              setDigestWebhook(checked);
              updatePreference({ DigestWebhook: checked });
            }}
            onWebhookURLChange={(value) => {
              setWebhookURL(value);
            }}
            onWebhookURLSave={() => {
              updatePreference({ WebhookURL: webhookURL });
            }}
            weeklySummary={weeklySummary}
            onToggleWeeklySummary={(checked) => {
              setWeeklySummary(checked);
              updatePreference({ WeeklySummary: checked });
            }}
          />
        </section>

        <section id="quiet-hours" className="scroll-mt-24">
          <QuietHoursSettings
            enabled={quietHoursEnabled}
            start={quietHoursStart}
            end={quietHoursEnd}
            prefLoading={prefLoading}
            prefSaving={prefSaving}
            onEnabledChange={(checked) => {
              setQuietHoursEnabled(checked);
              updatePreference({ QuietHoursEnabled: checked });
            }}
            onStartChange={(value) => {
              setQuietHoursStart(value);
              updatePreference({ QuietHoursStart: value });
            }}
            onEndChange={(value) => {
              setQuietHoursEnd(value);
              updatePreference({ QuietHoursEnd: value });
            }}
          />
        </section>

        <section id="digest-style" className="scroll-mt-24">
          <DigestStyleSettings
            digestStyle={digestStyle}
            prefLoading={prefLoading}
            prefSaving={prefSaving}
            onDigestStyleChange={(value) => {
              setDigestStyle(value);
              updatePreference({ DigestStyle: value });
            }}
          />
        </section>

        <section id="categories" className="scroll-mt-24">
          <DigestCategoriesSettings
            excludedCategories={excludedCategories}
            prefLoading={prefLoading}
            prefSaving={prefSaving}
            onExcludedCategoriesChange={(categories) => {
              setExcludedCategories(categories);
              updatePreference({ ExcludedCategories: categories });
            }}
          />
        </section>

        <section id="keywords" className="scroll-mt-24">
          <KeywordPreferencesSettings
            interestKeywords={interestKeywords}
            exclusionKeywords={exclusionKeywords}
            prefLoading={prefLoading}
            prefSaving={prefSaving}
            onInterestKeywordsChange={(keywords) => {
              setInterestKeywords(keywords);
              updatePreference({ InterestKeywords: keywords });
            }}
            onExclusionKeywordsChange={(keywords) => {
              setExclusionKeywords(keywords);
              updatePreference({ ExclusionKeywords: keywords });
            }}
          />
        </section>

        <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground pt-4">Organisation</h2>

        <section id="inbox" className="scroll-mt-24">
          <InboxDisplaySettings
            topicLimit={topicLimit}
            prefLoading={prefLoading}
            prefSaving={prefSaving}
            onTopicLimitChange={(value) => {
              setTopicLimit(value);
              updatePreference({ TopicLimit: value });
            }}
          />
        </section>

        <section id="labels" className="scroll-mt-24">
          <LabelSettings />
        </section>

        <section id="auto-archive" className="scroll-mt-24">
          <AutoArchiveSettings />
        </section>

        <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground pt-4">Advanced</h2>

        {accounts.length > 0 && (
          <section id="forwarding" className="scroll-mt-24">
            <Card>
              <CardHeader>
                <CardTitle>Forwarding Setup</CardTitle>
                <CardDescription>
                  Learn how to forward newsletters from your email provider
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ForwardingGuide emailAddress={accounts[0]?.EmailAddress} />
              </CardContent>
            </Card>
          </section>
        )}

        <section id="api-keys" className="scroll-mt-24">
          <ApiKeysSettings />
        </section>

        <section id="limits" className="scroll-mt-24">
          <DailyLimitSettings
            dailyTokenLimit={dailyTokenLimit}
            prefLoading={prefLoading}
            prefSaving={prefSaving}
            onDailyTokenLimitChange={(value) => {
              setDailyTokenLimit(value);
              updatePreference({ DailyTokenLimit: value });
            }}
          />
        </section>

        <section id="usage" className="scroll-mt-24">
          <UsageStatsCard />
        </section>

        <section id="data" className="scroll-mt-24">
          <Card>
            <CardHeader>
              <CardTitle>Your Data</CardTitle>
              <CardDescription>Export all your emails and digests as JSON</CardDescription>
            </CardHeader>
            <CardContent>
              <ExportDataButton />
            </CardContent>
          </Card>
        </section>

        <section id="account" className="scroll-mt-24">
          <Card>
            <CardHeader>
              <CardTitle>Account</CardTitle>
              <CardDescription>Manage your account settings</CardDescription>
            </CardHeader>
            <CardContent>
              <SignOutButton />
            </CardContent>
          </Card>
        </section>
      </div>
    </div>
  );
}
