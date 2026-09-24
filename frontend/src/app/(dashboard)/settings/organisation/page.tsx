"use client";

import { InboxDisplaySettings } from "../inbox-display-settings";
import { LabelSettings } from "../label-settings";
import { AutoArchiveSettings } from "../auto-archive-settings";
import { SettingsHeader, SettingsPanel } from "../settings-panel";
import { settingsGroupBySlug } from "../settings-groups";
import { usePreferences } from "../use-preferences";

const group = settingsGroupBySlug("organisation")!;

export default function OrganisationSettingsPage() {
  const { pref, setPref, prefLoading, prefSaving, updatePreference } = usePreferences();

  return (
    <>
      <SettingsHeader title={group.label} description={group.description} />

      <SettingsPanel>
        <InboxDisplaySettings
          topicLimit={pref.TopicLimit}
          prefLoading={prefLoading}
          prefSaving={prefSaving}
          onTopicLimitChange={(value) => {
            setPref((p) => ({ ...p, TopicLimit: value }));
            updatePreference({ TopicLimit: value });
          }}
        />
      </SettingsPanel>

      <section id="labels" className="scroll-mt-24">
        <LabelSettings />
      </section>

      <section id="auto-archive" className="scroll-mt-24">
        <AutoArchiveSettings />
      </section>
    </>
  );
}
