"use client";

import { Button } from "@/components/ui/button";
import { SettingRow } from "./settings-panel";

interface DigestStyleSettingsProps {
  digestStyle: string;
  prefLoading: boolean;
  prefSaving: boolean;
  onDigestStyleChange: (value: string) => void;
}

export function DigestStyleSettings({
  digestStyle,
  prefLoading,
  prefSaving,
  onDigestStyleChange,
}: DigestStyleSettingsProps) {
  return (
    <SettingRow
      id="digest-style"
      title="Digest style"
      description={
        digestStyle === "concise"
          ? "Shorter summaries with fewer highlights, for a quick overview"
          : "Full summaries with detailed highlights and more context"
      }
      control={
        <div className="flex gap-2">
          <Button
            type="button"
            size="sm"
            variant={digestStyle === "detailed" ? "default" : "outline"}
            onClick={() => onDigestStyleChange("detailed")}
            disabled={prefLoading || prefSaving}
          >
            Detailed
          </Button>
          <Button
            type="button"
            size="sm"
            variant={digestStyle === "concise" ? "default" : "outline"}
            onClick={() => onDigestStyleChange("concise")}
            disabled={prefLoading || prefSaving}
          >
            Concise
          </Button>
        </div>
      }
    />
  );
}
