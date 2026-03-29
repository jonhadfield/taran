"use client";

import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
    <Card>
      <CardHeader>
        <CardTitle>Daily Token Limit</CardTitle>
        <CardDescription>
          Limit daily AI processing to prevent burst usage. Emails exceeding the
          daily limit are deferred to the next day.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <Label htmlFor="daily-limit" className="flex flex-col items-start gap-1">
            <span>Enable daily limit</span>
            <span className="text-sm font-normal text-muted-foreground">
              Spread token usage evenly across the month
            </span>
          </Label>
          <Switch
            id="daily-limit"
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
        </div>

        {enabled && (
          <div className="space-y-3 border-t pt-4">
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
      </CardContent>
    </Card>
  );
}
