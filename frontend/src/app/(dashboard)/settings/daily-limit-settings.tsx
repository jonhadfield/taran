"use client";

import { useState } from "react";
import { SettingRow } from "./settings-panel";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

interface DailyLimitSettingsProps {
  dailyTokenLimit: number;
  prefLoading: boolean;
  prefSaving: boolean;
  onDailyTokenLimitChange: (value: number) => void;
}

const PRESET_LIMITS = [
  { value: 0, label: "No limit" },
  { value: 10000, label: "10K" },
  { value: 25000, label: "25K" },
  { value: 50000, label: "50K" },
  { value: 100000, label: "100K" },
];

export function DailyLimitSettings({
  dailyTokenLimit,
  prefLoading,
  prefSaving,
  onDailyTokenLimitChange,
}: DailyLimitSettingsProps) {
  const enabled = dailyTokenLimit > 0;
  const [customValue, setCustomValue] = useState(
    dailyTokenLimit > 0 ? dailyTokenLimit : 50000
  );

  return (
    <SettingRow
      id="limits"
      title="Daily AI limit"
      description="Spread usage across the month. Emails over the limit are analysed the next day."
      control={
        <Switch
          id="daily-limit"
          aria-label="Enable daily limit"
          checked={enabled}
          onCheckedChange={(checked) => {
              if (checked) {
                onDailyTokenLimitChange(customValue);
              } else {
                onDailyTokenLimitChange(0);
              }
          }}
          disabled={prefLoading || prefSaving}
        />
      }
    >
      {enabled && (
        <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
              {PRESET_LIMITS.filter((p) => p.value > 0).map((preset) => (
                <button
                  key={preset.value}
                  onClick={() => {
                    setCustomValue(preset.value);
                    onDailyTokenLimitChange(preset.value);
                  }}
                  disabled={prefSaving}
                  className={`rounded-md border px-3 py-1.5 text-sm transition-colors ${
                    dailyTokenLimit === preset.value
                      ? "bg-primary text-primary-foreground border-primary"
                      : "hover:bg-accent"
                  }`}
                >
                  {preset.label}
                </button>
              ))}
            </div>
            <div className="flex items-center gap-2">
              <Label htmlFor="custom-daily-limit" className="text-sm whitespace-nowrap">
                Custom:
              </Label>
              <input
                id="custom-daily-limit"
                type="number"
                min={1000}
                max={1000000}
                step={1000}
                value={customValue}
                onChange={(e) => setCustomValue(Number(e.target.value))}
                onBlur={() => {
                  if (customValue >= 1000) {
                    onDailyTokenLimitChange(customValue);
                  }
                }}
                disabled={prefSaving}
                className="flex h-9 w-28 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
            <span className="text-sm text-muted-foreground">tokens/day</span>
          </div>
        </div>
      )}
    </SettingRow>
  );
}
