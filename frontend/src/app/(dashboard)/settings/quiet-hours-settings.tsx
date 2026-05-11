"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { HOUR_OPTIONS } from "@/lib/constants";
import { NativeSelect } from "@/components/ui/native-select";

interface QuietHoursSettingsProps {
  enabled: boolean;
  start: number;
  end: number;
  prefLoading: boolean;
  prefSaving: boolean;
  onEnabledChange: (checked: boolean) => void;
  onStartChange: (value: number) => void;
  onEndChange: (value: number) => void;
}

export function QuietHoursSettings({
  enabled,
  start,
  end,
  prefLoading,
  prefSaving,
  onEnabledChange,
  onStartChange,
  onEndChange,
}: QuietHoursSettingsProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Quiet Hours</CardTitle>
        <CardDescription>
          Defer email processing during specific hours
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="flex items-center justify-between">
          <Label htmlFor="quiet-hours" className="flex flex-col items-start gap-1">
            <span>Enable quiet hours</span>
            <span className="text-sm font-normal text-muted-foreground">
              Emails received during this window are stored but processing is deferred
            </span>
          </Label>
          <Switch
            id="quiet-hours"
            checked={enabled}
            onCheckedChange={onEnabledChange}
            disabled={prefLoading || prefSaving}
          />
        </div>

        {enabled && (
          <div className="space-y-4 border-t pt-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="quiet-start">From</Label>
                <NativeSelect
                  id="quiet-start"
                  value={start}
                  onChange={(e) => onStartChange(Number(e.target.value))}
                  disabled={prefSaving}
                  wrapperClassName="w-full"
                  className="transition-colors"
                >
                  {HOUR_OPTIONS.map(({ value, label }) => (
                    <option key={value} value={value}>{label}</option>
                  ))}
                </NativeSelect>
              </div>
              <div className="space-y-2">
                <Label htmlFor="quiet-end">Until</Label>
                <NativeSelect
                  id="quiet-end"
                  value={end}
                  onChange={(e) => onEndChange(Number(e.target.value))}
                  disabled={prefSaving}
                  wrapperClassName="w-full"
                  className="transition-colors"
                >
                  {HOUR_OPTIONS.map(({ value, label }) => (
                    <option key={value} value={value}>{label}</option>
                  ))}
                </NativeSelect>
              </div>
            </div>
            <p className="text-xs text-muted-foreground">
              Uses your digest timezone setting. Emails will be processed once quiet hours end.
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
