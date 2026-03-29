"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
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

const DEFAULT_PREF: UserPreference = {
    UserID: "", DigestEmail: false, DigestFrequency: "daily", DigestHour: 7,
    DigestDay: 1, DigestTimezone: "UTC", TopicLimit: 15, DigestStyle: "detailed",
    InterestKeywords: [], ExclusionKeywords: [], ColorTheme: "neutral",
    ExcludedCategories: ["notification", "transactional", "marketing"],
    DailyTokenLimit: 0, QuietHoursEnabled: false, QuietHoursStart: 22,
    QuietHoursEnd: 7, WeeklySummary: true, DigestWebhook: false, WebhookURL: "",
    CreatedAt: "", UpdatedAt: "",
};

export default function SettingsPage() {
  const { colorTheme, setColorTheme } = useColorTheme();
  const [accounts, setAccounts] = useState<EmailAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [pref, setPref] = useState<UserPreference>(DEFAULT_PREF);
  const [prefLoading, setPrefLoading] = useState(true);
  const [prefSaving, setPrefSaving] = useState(false);

  // Convenience aliases used by child components
  const digestEmail = pref.DigestEmail;
  const digestFrequency = pref.DigestFrequency;
  const digestHour = pref.DigestHour;
  const digestDay = pref.DigestDay;
  const digestTimezone = pref.DigestTimezone;
  const digestStyle = pref.DigestStyle;
  const interestKeywords = pref.InterestKeywords || [];
  const exclusionKeywords = pref.ExclusionKeywords || [];
  const excludedCategories = pref.ExcludedCategories || ["notification", "transactional", "marketing"];
  const topicLimit = pref.TopicLimit;
  const dailyTokenLimit = pref.DailyTokenLimit;
  const quietHoursEnabled = pref.QuietHoursEnabled;
  const quietHoursStart = pref.QuietHoursStart;
  const quietHoursEnd = pref.QuietHoursEnd;
  const digestWebhook = pref.DigestWebhook;
  const webhookURL = pref.WebhookURL;
  const weeklySummary = pref.WeeklySummary;

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
      const data = await apiGet<UserPreference>("preferences");
      // Apply defaults for missing fields
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
    fetchAccounts();
    fetchPreferences();
  }, [fetchAccounts, fetchPreferences]);

  const updatePreference = async (updates: Partial<UserPreference>) => {
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

  // Build timezone list: user's detected TZ first, then common ones (stable across renders)
  const timezoneOptions = useMemo(() => {
    const tz = detectTimezone();
    return COMMON_TIMEZONES.includes(tz) ? COMMON_TIMEZONES : [tz, ...COMMON_TIMEZONES];
  }, []);

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
              setPref(p => ({...p, DigestEmail: checked}));
              updatePreference({ DigestEmail: checked });
            }}
            onFrequencyChange={(value) => {
              setPref(p => ({...p, DigestFrequency: value as "daily" | "weekly"}));
              updatePreference({ DigestFrequency: value as "daily" | "weekly" });
            }}
            onHourChange={(value) => {
              setPref(p => ({...p, DigestHour: value}));
              updatePreference({ DigestHour: value });
            }}
            onDayChange={(value) => {
              setPref(p => ({...p, DigestDay: value}));
              updatePreference({ DigestDay: value });
            }}
            onTimezoneChange={(value) => {
              setPref(p => ({...p, DigestTimezone: value}));
              updatePreference({ DigestTimezone: value });
            }}
            digestWebhook={digestWebhook}
            webhookURL={webhookURL}
            onToggleDigestWebhook={(checked) => {
              setPref(p => ({...p, DigestWebhook: checked}));
              updatePreference({ DigestWebhook: checked });
            }}
            onWebhookURLChange={(value) => {
              setPref(p => ({...p, WebhookURL: value}));
            }}
            onWebhookURLSave={() => {
              updatePreference({ WebhookURL: webhookURL });
            }}
            weeklySummary={weeklySummary}
            onToggleWeeklySummary={(checked) => {
              setPref(p => ({...p, WeeklySummary: checked}));
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
              setPref(p => ({...p, QuietHoursEnabled: checked}));
              updatePreference({ QuietHoursEnabled: checked });
            }}
            onStartChange={(value) => {
              setPref(p => ({...p, QuietHoursStart: value}));
              updatePreference({ QuietHoursStart: value });
            }}
            onEndChange={(value) => {
              setPref(p => ({...p, QuietHoursEnd: value}));
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
              setPref(p => ({...p, DigestStyle: value as "detailed" | "concise"}));
              updatePreference({ DigestStyle: value as "detailed" | "concise" });
            }}
          />
        </section>

        <section id="categories" className="scroll-mt-24">
          <DigestCategoriesSettings
            excludedCategories={excludedCategories}
            prefLoading={prefLoading}
            prefSaving={prefSaving}
            onExcludedCategoriesChange={(categories) => {
              setPref(p => ({...p, ExcludedCategories: categories}));
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
              setPref(p => ({...p, InterestKeywords: keywords}));
              updatePreference({ InterestKeywords: keywords });
            }}
            onExclusionKeywordsChange={(keywords) => {
              setPref(p => ({...p, ExclusionKeywords: keywords}));
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
              setPref(p => ({...p, TopicLimit: value}));
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
              setPref(p => ({...p, DailyTokenLimit: value}));
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
