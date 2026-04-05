"use client";

import { useState } from "react";
import { toast } from "sonner";
import { apiPost } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { HOUR_OPTIONS, nativeSelectClassName } from "@/lib/constants";
import { cn } from "@/lib/utils";

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
            onCheckedChange={onToggleDigestEmail}
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
                  <select
                    id="digest-day"
                    value={digestDay}
                    onChange={(e) => onDayChange(Number(e.target.value))}
                    disabled={prefSaving}
                    className={cn(nativeSelectClassName, "flex w-full max-w-full sm:max-w-xs py-1 transition-colors")}
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
                onChange={(e) => onHourChange(Number(e.target.value))}
                disabled={prefSaving}
                className={cn(nativeSelectClassName, "flex w-full max-w-full sm:max-w-xs py-1 transition-colors")}
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
                onChange={(e) => onTimezoneChange(e.target.value)}
                disabled={prefSaving}
                className={cn(nativeSelectClassName, "flex w-full max-w-full sm:max-w-xs py-1 transition-colors")}
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
        <Separator className="my-2" />

        <div className="flex items-center justify-between">
          <Label htmlFor="digest-webhook" className="flex flex-col items-start gap-1">
            <span>Webhook delivery</span>
            <span className="text-sm font-normal text-muted-foreground">
              POST digest summaries to a URL (Slack, Zapier, etc.)
            </span>
          </Label>
          <Switch
            id="digest-webhook"
            checked={digestWebhook}
            onCheckedChange={onToggleDigestWebhook}
            disabled={prefLoading || prefSaving}
          />
        </div>

        {digestWebhook && (
          <div className="space-y-2 border-t pt-4">
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
              Digest summaries will be POSTed as JSON with title, summary, highlights, and a link to view the full digest.
            </p>
          </div>
        )}

      </CardContent>
    </Card>
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
