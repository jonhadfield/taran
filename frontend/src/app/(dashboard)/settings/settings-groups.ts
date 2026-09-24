/**
 * Settings are split into four pages. Grouping them by the job someone came to
 * do keeps each page short enough to take in at a glance, and keeps the
 * navigation visible instead of scrolling sideways.
 */
export interface SettingsGroup {
  slug: string;
  label: string;
  description: string;
  /** Section anchors within the page, in the order they appear. */
  sections: { id: string; label: string }[];
}

export const SETTINGS_GROUPS: SettingsGroup[] = [
  {
    slug: "inbox",
    label: "Inbox",
    description: "Your MailBrief address, how to forward mail to it, and newsletters to try",
    sections: [
      { id: "accounts", label: "Address" },
      { id: "forwarding", label: "Forwarding" },
      { id: "newsletters", label: "Newsletters" },
    ],
  },
  {
    slug: "digest",
    label: "Digest",
    description: "When your digest arrives and what goes into it",
    sections: [
      { id: "delivery", label: "Delivery" },
      { id: "quiet-hours", label: "Quiet hours" },
      { id: "digest-style", label: "Style" },
      { id: "categories", label: "Categories" },
      { id: "keywords", label: "Keywords" },
      { id: "analysis-rules", label: "Analysis rules" },
    ],
  },
  {
    slug: "organisation",
    label: "Organisation",
    description: "How your inbox is displayed, labelled and tidied up",
    sections: [
      { id: "inbox-display", label: "Inbox display" },
      { id: "labels", label: "Labels" },
      { id: "auto-archive", label: "Auto-archive" },
    ],
  },
  {
    slug: "account",
    label: "Account",
    description: "Appearance, AI keys, usage limits and your data",
    sections: [
      { id: "appearance", label: "Appearance" },
      { id: "api-keys", label: "AI keys" },
      { id: "limits", label: "Limits" },
      { id: "usage", label: "Usage" },
      { id: "data", label: "Your data" },
      { id: "account", label: "Sign out" },
    ],
  },
];

export function settingsGroupBySlug(slug: string): SettingsGroup | undefined {
  return SETTINGS_GROUPS.find((g) => g.slug === slug);
}
