"use client";

import { Switch } from "@/components/ui/switch";
import { SettingRow } from "./settings-panel";

interface DigestSectionsSettingsProps {
  includeHighlights: boolean;
  includeTopics: boolean;
  includeEmails: boolean;
  prefLoading: boolean;
  prefSaving: boolean;
  onIncludeHighlightsChange: (value: boolean) => void;
  onIncludeTopicsChange: (value: boolean) => void;
  onIncludeEmailsChange: (value: boolean) => void;
}

export function DigestSectionsSettings({
  includeHighlights,
  includeTopics,
  includeEmails,
  prefLoading,
  prefSaving,
  onIncludeHighlightsChange,
  onIncludeTopicsChange,
  onIncludeEmailsChange,
}: DigestSectionsSettingsProps) {
  const disabled = prefLoading || prefSaving;

  return (
    <>
      <SettingRow
        id="digest-sections"
        title="Highlights"
        description="Key takeaways in each digest. The summary is always included."
        control={
          <Switch
            checked={includeHighlights}
            onCheckedChange={onIncludeHighlightsChange}
            disabled={disabled}
            aria-label="Include highlights"
          />
        }
      />
      <SettingRow
        id="digest-section-topics"
        title="Topics"
        description="Top topic tags for the period"
        control={
          <Switch
            checked={includeTopics}
            onCheckedChange={onIncludeTopicsChange}
            disabled={disabled}
            aria-label="Include topics"
          />
        }
      />
      <SettingRow
        id="digest-section-emails"
        title="Included emails"
        description="List of emails that went into the digest"
        control={
          <Switch
            checked={includeEmails}
            onCheckedChange={onIncludeEmailsChange}
            disabled={disabled}
            aria-label="Include email list"
          />
        }
      />
    </>
  );
}
