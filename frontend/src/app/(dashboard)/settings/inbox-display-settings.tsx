"use client";

import { SettingRow } from "./settings-panel";

interface InboxDisplaySettingsProps {
  topicLimit: number;
  prefLoading: boolean;
  prefSaving: boolean;
  onTopicLimitChange: (value: number) => void;
}

export function InboxDisplaySettings({
  topicLimit,
  prefLoading,
  prefSaving,
  onTopicLimitChange,
}: InboxDisplaySettingsProps) {
  return (
    <SettingRow
      id="inbox-display"
      title="Topic cloud size"
      description="Maximum number of topics shown in your inbox word cloud"
      control={
        <div className="flex items-center gap-3">
          <input
            id="topic-limit"
            aria-label="Topic cloud size"
            type="range"
            min={5}
            max={50}
            step={5}
            value={topicLimit}
            onChange={(e) => onTopicLimitChange(Number(e.target.value))}
            disabled={prefLoading || prefSaving}
            className="w-40 accent-primary"
          />
          <span className="w-8 text-right text-sm font-medium">{topicLimit}</span>
        </div>
      }
    />
  );
}
