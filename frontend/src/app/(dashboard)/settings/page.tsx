"use client";

import { useEffect, useState } from "react";
import { apiGet, apiDelete, apiPatch } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Loader2, Trash2 } from "lucide-react";
import type { EmailAccount, ListResponse, UserPreference } from "@/types/api";
import { CopyButton } from "./copy-button";
import { SignOutButton } from "./sign-out-button";
import { UsernameForm } from "@/components/username-form";
import { ForwardingGuide } from "@/components/forwarding-guide";

const HOUR_OPTIONS = Array.from({ length: 24 }, (_, i) => {
  const hour = i % 12 || 12;
  const ampm = i < 12 ? "AM" : "PM";
  return { value: i, label: `${hour}:00 ${ampm}` };
});

const DAY_OPTIONS = [
  { value: 0, label: "Sunday" },
  { value: 1, label: "Monday" },
  { value: 2, label: "Tuesday" },
  { value: 3, label: "Wednesday" },
  { value: 4, label: "Thursday" },
  { value: 5, label: "Friday" },
  { value: 6, label: "Saturday" },
];

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

  const updatePreference = async (updates: Partial<Pick<UserPreference, "DigestEmail" | "DigestFrequency" | "DigestHour" | "DigestDay" | "DigestTimezone" | "TopicLimit">>) => {
    setPrefSaving(true);
    try {
      const updated = await apiPatch<UserPreference>("preferences", updates);
      setDigestEmail(updated.DigestEmail);
      setDigestFrequency(updated.DigestFrequency || "daily");
      setDigestHour(updated.DigestHour ?? 7);
      setDigestDay(updated.DigestDay ?? 1);
      setDigestTimezone(updated.DigestTimezone || "UTC");
      setTopicLimit(updated.TopicLimit ?? 15);
    } catch {
      // Revert on error by re-fetching
      await fetchPreferences();
    } finally {
      setPrefSaving(false);
    }
  };

  const handleToggleDigestEmail = async (checked: boolean) => {
    setDigestEmail(checked);
    await updatePreference({ DigestEmail: checked });
  };

  const handleFrequencyChange = async (value: string) => {
    setDigestFrequency(value);
    await updatePreference({ DigestFrequency: value });
  };

  const handleHourChange = async (value: number) => {
    setDigestHour(value);
    await updatePreference({ DigestHour: value });
  };

  const handleDayChange = async (value: number) => {
    setDigestDay(value);
    await updatePreference({ DigestDay: value });
  };

  const handleTimezoneChange = async (value: string) => {
    setDigestTimezone(value);
    await updatePreference({ DigestTimezone: value });
  };

  const handleTopicLimitChange = async (value: number) => {
    setTopicLimit(value);
    await updatePreference({ TopicLimit: value });
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);

    try {
      await apiDelete(`accounts/${deleteTarget.ID}`);
      setDeleteTarget(null);
      await fetchAccounts();
    } catch {
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

      <Card>
        <CardHeader>
          <CardTitle>Email Accounts</CardTitle>
          <CardDescription>
            Your managed inboxes for receiving newsletters
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {loading && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              Loading accounts...
            </div>
          )}

          {error && <p className="text-sm text-destructive">{error}</p>}

          {!loading && accounts.length === 0 && !error && (
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground">
                You haven&apos;t set up an inbox yet. Choose a username to get started.
              </p>
              <UsernameForm onSuccess={fetchAccounts} />
            </div>
          )}

          {accounts.map((account) => (
            <div
              key={account.ID}
              className="flex items-center justify-between rounded-lg border p-4"
            >
              <div>
                <p className="font-medium">
                  {account.DisplayName || "My Inbox"}
                </p>
                <p className="text-sm text-muted-foreground">
                  {account.EmailAddress}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <CopyButton text={account.EmailAddress} />
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setDeleteTarget(account)}
                >
                  <Trash2 className="h-4 w-4 text-muted-foreground" />
                </Button>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Digest Delivery</CardTitle>
          <CardDescription>
            Configure how and when you receive your digest
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between">
            <Label htmlFor="digest-email" className="flex flex-col items-start gap-1">
              <span>Email delivery</span>
              <span className="text-sm font-normal text-muted-foreground">
                Receive your digest as an email
              </span>
            </Label>
            <Switch
              id="digest-email"
              checked={digestEmail}
              onCheckedChange={handleToggleDigestEmail}
              disabled={prefLoading || prefSaving}
            />
          </div>

          {digestEmail && (
            <div className="space-y-4 border-t pt-4">
              <div className="space-y-2">
                <Label htmlFor="digest-frequency">Frequency</Label>
                <div className="flex gap-2">
                  <Button
                    type="button"
                    size="sm"
                    variant={digestFrequency === "daily" ? "default" : "outline"}
                    onClick={() => handleFrequencyChange("daily")}
                    disabled={prefSaving}
                  >
                    Daily
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant={digestFrequency === "weekly" ? "default" : "outline"}
                    onClick={() => handleFrequencyChange("weekly")}
                    disabled={prefSaving}
                  >
                    Weekly
                  </Button>
                </div>
                {digestFrequency === "weekly" && (
                  <div className="space-y-1">
                    <Label htmlFor="digest-day">Day of week</Label>
                    <select
                      id="digest-day"
                      value={digestDay}
                      onChange={(e) => handleDayChange(Number(e.target.value))}
                      disabled={prefSaving}
                      className="flex h-9 w-full max-w-full sm:max-w-xs rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                    >
                      {DAY_OPTIONS.map(({ value, label }) => (
                        <option key={value} value={value}>{label}</option>
                      ))}
                    </select>
                  </div>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="digest-hour">Delivery time</Label>
                <select
                  id="digest-hour"
                  value={digestHour}
                  onChange={(e) => handleHourChange(Number(e.target.value))}
                  disabled={prefSaving}
                  className="flex h-9 w-full max-w-full sm:max-w-xs rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                >
                  {HOUR_OPTIONS.map(({ value, label }) => (
                    <option key={value} value={value}>{label}</option>
                  ))}
                </select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="digest-timezone">Timezone</Label>
                <select
                  id="digest-timezone"
                  value={digestTimezone}
                  onChange={(e) => handleTimezoneChange(e.target.value)}
                  disabled={prefSaving}
                  className="flex h-9 w-full max-w-full sm:max-w-xs rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                >
                  {timezoneOptions.map((tz) => (
                    <option key={tz} value={tz}>
                      {tz.replace(/_/g, " ")}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Inbox Display</CardTitle>
          <CardDescription>
            Customize how your inbox looks
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="topic-limit">Topic cloud size</Label>
            <p className="text-sm text-muted-foreground">
              Maximum number of topics shown in your inbox word cloud
            </p>
            <div className="flex items-center gap-3">
              <input
                id="topic-limit"
                type="range"
                min={5}
                max={50}
                step={5}
                value={topicLimit}
                onChange={(e) => handleTopicLimitChange(Number(e.target.value))}
                disabled={prefLoading || prefSaving}
                className="w-full max-w-xs accent-primary"
              />
              <span className="text-sm font-medium w-8 text-right">{topicLimit}</span>
            </div>
          </div>
        </CardContent>
      </Card>

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

      <Dialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Inbox</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete{" "}
              <span className="font-medium">
                {deleteTarget?.EmailAddress}
              </span>
              ? This inbox will stop receiving emails.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleting}
            >
              {deleting && <Loader2 className="mr-1 h-4 w-4 animate-spin" />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
