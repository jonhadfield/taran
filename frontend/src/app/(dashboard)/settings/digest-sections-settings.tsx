"use client";

import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { ChevronDown, ChevronUp } from "lucide-react";
import { SettingRow } from "./settings-panel";
import {
  type DigestDensity,
  type DigestSectionId,
  moveDigestSection,
  normalizeDigestSectionOrder,
} from "@/lib/digest-template";

const SECTION_META: Record<
  DigestSectionId,
  { title: string; description: string; aria: string; includeKey: "highlights" | "topics" | "emails" }
> = {
  highlights: {
    title: "Highlights",
    description: "Key takeaways in each digest. The summary is always first.",
    aria: "Include highlights",
    includeKey: "highlights",
  },
  topics: {
    title: "Topics",
    description: "Top topic tags for the period",
    aria: "Include topics",
    includeKey: "topics",
  },
  emails: {
    title: "Included emails",
    description: "List of emails that went into the digest",
    aria: "Include email list",
    includeKey: "emails",
  },
};

interface DigestSectionsSettingsProps {
  includeHighlights: boolean;
  includeTopics: boolean;
  includeEmails: boolean;
  sectionOrder: DigestSectionId[];
  density: DigestDensity;
  prefLoading: boolean;
  prefSaving: boolean;
  onIncludeHighlightsChange: (value: boolean) => void;
  onIncludeTopicsChange: (value: boolean) => void;
  onIncludeEmailsChange: (value: boolean) => void;
  onSectionOrderChange: (order: DigestSectionId[]) => void;
  onDensityChange: (value: DigestDensity) => void;
}

export function DigestSectionsSettings({
  includeHighlights,
  includeTopics,
  includeEmails,
  sectionOrder,
  density,
  prefLoading,
  prefSaving,
  onIncludeHighlightsChange,
  onIncludeTopicsChange,
  onIncludeEmailsChange,
  onSectionOrderChange,
  onDensityChange,
}: DigestSectionsSettingsProps) {
  const disabled = prefLoading || prefSaving;
  const order = normalizeDigestSectionOrder(sectionOrder);
  const included = {
    highlights: includeHighlights,
    topics: includeTopics,
    emails: includeEmails,
  };
  const onInclude = {
    highlights: onIncludeHighlightsChange,
    topics: onIncludeTopicsChange,
    emails: onIncludeEmailsChange,
  };

  return (
    <>
      <SettingRow
        id="digest-density"
        title="Digest density"
        description={
          density === "compact"
            ? "Tighter spacing in the digest view, email, and PDF"
            : "Roomier spacing between digest sections"
        }
        control={
          <div className="flex gap-2">
            <Button
              type="button"
              size="sm"
              variant={density === "comfortable" ? "default" : "outline"}
              onClick={() => onDensityChange("comfortable")}
              disabled={disabled}
            >
              Comfortable
            </Button>
            <Button
              type="button"
              size="sm"
              variant={density === "compact" ? "default" : "outline"}
              onClick={() => onDensityChange("compact")}
              disabled={disabled}
            >
              Compact
            </Button>
          </div>
        }
      />

      {order.map((id, index) => {
        const meta = SECTION_META[id];
        return (
          <SettingRow
            key={id}
            id={index === 0 ? "digest-sections" : `digest-section-${id}`}
            title={meta.title}
            description={meta.description}
            control={
              <div className="flex items-center gap-2">
                <div className="flex flex-col">
                  <Button
                    type="button"
                    size="icon-sm"
                    variant="ghost"
                    aria-label={`Move ${meta.title} up`}
                    disabled={disabled || index === 0}
                    onClick={() =>
                      onSectionOrderChange(moveDigestSection(order, index, -1))
                    }
                  >
                    <ChevronUp className="size-4" />
                  </Button>
                  <Button
                    type="button"
                    size="icon-sm"
                    variant="ghost"
                    aria-label={`Move ${meta.title} down`}
                    disabled={disabled || index === order.length - 1}
                    onClick={() =>
                      onSectionOrderChange(moveDigestSection(order, index, 1))
                    }
                  >
                    <ChevronDown className="size-4" />
                  </Button>
                </div>
                <Switch
                  checked={included[meta.includeKey]}
                  onCheckedChange={onInclude[meta.includeKey]}
                  disabled={disabled}
                  aria-label={meta.aria}
                />
              </div>
            }
          />
        );
      })}
    </>
  );
}
