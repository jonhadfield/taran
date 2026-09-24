"use client";

import { SettingRow } from "./settings-panel";
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
    <SettingRow
      id="quiet-hours"
      title="Quiet hours"
      description="Emails arriving in this window are stored, and analysed once it ends"
      control={
        <Switch
          id="quiet-hours"
          aria-label="Enable quiet hours"
          checked={enabled}
          onCheckedChange={onEnabledChange}
          disabled={prefLoading || prefSaving}
        />
      }
    >
      {enabled && (
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4 sm:max-w-sm">
            <div className="space-y-2">
              <Label htmlFor="quiet-start">From</Label>
                <NativeSelect
                  id="quiet-start"
                  value={start}
                  onChange={(e) => onStartChange(Number(e.target.value))}
                  disabled={prefSaving}
                  wrapperClassName="w-full"
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
                >
                  {HOUR_OPTIONS.map(({ value, label }) => (
                    <option key={value} value={value}>{label}</option>
                  ))}
                </NativeSelect>
              </div>
            </div>
          <p className="text-xs text-muted-foreground">
            Uses your digest timezone.
          </p>
        </div>
      )}
    </SettingRow>
  );
}
