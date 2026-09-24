"use client";

import { useState } from "react";
import { toast } from "sonner";
import { apiPost } from "@/lib/api";
import { SettingRow } from "./settings-panel";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { HOUR_OPTIONS } from "@/lib/constants";
import { NativeSelect } from "@/components/ui/native-select";

const DAY_OPTIONS = [
  { value: 0, label: "Sunday" },
  { value: 1, label: "Monday" },
  { value: 2, label: "Tuesday" },
  { value: 3, label: "Wednesday" },
  { value: 4, label: "Thursday" },
  { value: 5, label: "Friday" },
  { value: 6, label: "Saturday" },
];

interface DigestDeliverySettingsProps {
  digestEmail: boolean;
  digestFrequency: string;
  digestHour: number;
  digestDay: number;
  digestTimezone: string;
  timezoneOptions: string[];
  prefLoading: boolean;
  prefSaving: boolean;
  onToggleDigestEmail: (checked: boolean) => void;
  onFrequencyChange: (value: string) => void;
  onHourChange: (value: number) => void;
  onDayChange: (value: number) => void;
  onTimezoneChange: (value: string) => void;
  digestWebhook: boolean;
  webhookURL: string;
  onToggleDigestWebhook: (checked: boolean) => void;
  onWebhookURLChange: (value: string) => void;
  onWebhookURLSave: () => void;
}

export function DigestDeliverySettings({
  digestEmail,
  digestFrequency,
  digestHour,
  digestDay,
  digestTimezone,
  timezoneOptions,
  prefLoading,
  prefSaving,
  onToggleDigestEmail,
  onFrequencyChange,
  onHourChange,
  onDayChange,
  onTimezoneChange,
  digestWebhook,
  webhookURL,
  onToggleDigestWebhook,
  onWebhookURLChange,
  onWebhookURLSave,
}: DigestDeliverySettingsProps) {
  return (
    <>
      <SettingRow
        id="delivery"
        title="Email delivery"
        description="Receive your digest as an email"
        control={
          <Switch
            id="digest-email"
            aria-label="Email delivery"
            checked={digestEmail}
            onCheckedChange={onToggleDigestEmail}
            disabled={prefLoading || prefSaving}
          />
        }
      >
        {digestEmail && (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="digest-frequency">Frequency</Label>
              <div className="flex gap-2">
                <Button
                  type="button"
                  size="sm"
                  variant={digestFrequency === "daily" ? "default" : "outline"}
                  onClick={() => onFrequencyChange("daily")}
                  disabled={prefSaving}
                >
                  Daily
                </Button>
                <Button
                  type="button"
                  size="sm"
                  variant={digestFrequency === "weekly" ? "default" : "outline"}
                  onClick={() => onFrequencyChange("weekly")}
                  disabled={prefSaving}
                >
                  Weekly
                </Button>
              </div>
              {digestFrequency === "weekly" && (
                <div className="space-y-1">
                  <Label htmlFor="digest-day">Day of week</Label>
                  <NativeSelect
                    id="digest-day"
                    value={digestDay}
                    onChange={(e) => onDayChange(Number(e.target.value))}
                    disabled={prefSaving}
                    wrapperClassName="w-full sm:max-w-xs"
                  >
                    {DAY_OPTIONS.map(({ value, label }) => (
                      <option key={value} value={value}>{label}</option>
                    ))}
                  </NativeSelect>
                </div>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="digest-hour">Delivery time</Label>
              <NativeSelect
                id="digest-hour"
                value={digestHour}
                onChange={(e) => onHourChange(Number(e.target.value))}
                disabled={prefSaving}
                wrapperClassName="w-full sm:max-w-xs"
              >
                {HOUR_OPTIONS.map(({ value, label }) => (
                  <option key={value} value={value}>{label}</option>
                ))}
              </NativeSelect>
            </div>

            <div className="space-y-2">
              <Label htmlFor="digest-timezone">Timezone</Label>
              <NativeSelect
                id="digest-timezone"
                value={digestTimezone}
                onChange={(e) => onTimezoneChange(e.target.value)}
                disabled={prefSaving}
                wrapperClassName="w-full sm:max-w-xs"
              >
                {timezoneOptions.map((tz) => (
                  <option key={tz} value={tz}>
                    {tz.replace(/_/g, " ")}
                  </option>
                ))}
              </NativeSelect>
            </div>
          </div>
        )}
      </SettingRow>

      <SettingRow
        title="Webhook delivery"
        description="POST digest summaries to a URL (Slack, Zapier and the like)"
        control={
          <Switch
            id="digest-webhook"
            aria-label="Webhook delivery"
            checked={digestWebhook}
            onCheckedChange={onToggleDigestWebhook}
            disabled={prefLoading || prefSaving}
          />
        }
      >
        {digestWebhook && (
          <div className="space-y-2">
            <Label htmlFor="webhook-url">Webhook URL</Label>
            <div className="flex gap-2">
              <Input
                id="webhook-url"
                type="url"
                placeholder="https://hooks.slack.com/services/..."
                value={webhookURL}
                onChange={(e) => onWebhookURLChange(e.target.value)}
                disabled={prefSaving}
                className="flex-1"
              />
              <Button
                type="button"
                size="sm"
                onClick={onWebhookURLSave}
                disabled={prefSaving || !webhookURL}
              >
                Save
              </Button>
            </div>
            <WebhookTestButton disabled={prefSaving || !webhookURL} />
            <p className="text-xs text-muted-foreground">
              Sent as JSON with the title, summary, highlights and a link to the full digest.
            </p>
          </div>
        )}
      </SettingRow>
    </>
  );
}

function WebhookTestButton({ disabled }: { disabled: boolean }) {
  const [testing, setTesting] = useState(false);

  async function handleTest() {
    setTesting(true);
    try {
      await apiPost("preferences/test-webhook", {});
      toast.success("Test webhook sent successfully");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Webhook test failed");
    } finally {
      setTesting(false);
    }
  }

  return (
    <Button
      type="button"
      variant="outline"
      size="sm"
      onClick={handleTest}
      disabled={disabled || testing}
    >
      {testing ? "Sending..." : "Send test"}
    </Button>
  );
}
