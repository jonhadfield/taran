"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { apiGet, apiDelete, apiPatch } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import type { EmailAccount, ListResponse, UserPreference } from "@/types/api";
import { SignOutButton } from "./sign-out-button";
import { ForwardingGuide } from "@/components/forwarding-guide";
import { AccountSettings } from "./account-settings";
import { DigestDeliverySettings } from "./digest-delivery-settings";
import { InboxDisplaySettings } from "./inbox-display-settings";
import { DigestStyleSettings } from "./digest-style-settings";
import { KeywordPreferencesSettings } from "./keyword-preferences-settings";

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

export default function SettingsPage() {
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
  const [prefLoading, setPrefLoading] = useState(true);
  const [prefSaving, setPrefSaving] = useState(false);

  // Delete confirmation
  const [deleteTarget, setDeleteTarget] = useState<EmailAccount | null>(null);
  const [deleting, setDeleting] = useState(false);

  const fetchAccounts = async () => {
    try {
      const res = await apiGet<ListResponse<EmailAccount>>("accounts");
      setAccounts(res.data || []);
      setError("");
    } catch {
      setError("Failed to load accounts");
    } finally {
      setLoading(false);
    }
  };

  const fetchPreferences = async () => {
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
    } catch {
      setDigestTimezone(detectTimezone());
    } finally {
      setPrefLoading(false);
    }
  };

  useEffect(() => {
    fetchAccounts();
    fetchPreferences();
  }, []);

  const updatePreference = async (updates: Partial<Pick<UserPreference, "DigestEmail" | "DigestFrequency" | "DigestHour" | "DigestDay" | "DigestTimezone" | "TopicLimit" | "DigestStyle" | "InterestKeywords" | "ExclusionKeywords">>) => {
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

  return (
    <div className="space-y-6 max-w-2xl">
      <h1 className="text-2xl font-bold">Settings</h1>

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
      />

      <InboxDisplaySettings
        topicLimit={topicLimit}
        prefLoading={prefLoading}
        prefSaving={prefSaving}
        onTopicLimitChange={(value) => {
          setTopicLimit(value);
          updatePreference({ TopicLimit: value });
        }}
      />

      <DigestStyleSettings
        digestStyle={digestStyle}
        prefLoading={prefLoading}
        prefSaving={prefSaving}
        onDigestStyleChange={(value) => {
          setDigestStyle(value);
          updatePreference({ DigestStyle: value });
        }}
      />

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

      {accounts.length > 0 && (
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
      )}

      <Separator />

      <Card>
        <CardHeader>
          <CardTitle>Account</CardTitle>
          <CardDescription>Manage your account settings</CardDescription>
        </CardHeader>
        <CardContent>
          <SignOutButton />
        </CardContent>
      </Card>
    </div>
  );
}
