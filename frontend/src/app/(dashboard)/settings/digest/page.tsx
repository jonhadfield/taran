"use client";

import { DigestDeliverySettings } from "../digest-delivery-settings";
import { QuietHoursSettings } from "../quiet-hours-settings";
import { DigestStyleSettings } from "../digest-style-settings";
import { DigestCategoriesSettings } from "../digest-categories-settings";
import { KeywordPreferencesSettings } from "../keyword-preferences-settings";
import { AnalysisRulesSettings } from "../analysis-rules-settings";
import { SettingsHeader, SettingsPanel } from "../settings-panel";
import { settingsGroupBySlug } from "../settings-groups";
import { usePreferences } from "../use-preferences";

const group = settingsGroupBySlug("digest")!;

export default function DigestSettingsPage() {
  const { pref, setPref, prefLoading, prefSaving, updatePreference, handlePrefChange, timezoneOptions } =
    usePreferences();

  return (
    <>
      <SettingsHeader title={group.label} description={group.description} />

      <SettingsPanel>
        <DigestDeliverySettings
          digestEmail={pref.DigestEmail}
          digestFrequency={pref.DigestFrequency}
          digestHour={pref.DigestHour}
          digestDay={pref.DigestDay}
          digestTimezone={pref.DigestTimezone}
          timezoneOptions={timezoneOptions}
          prefLoading={prefLoading}
          prefSaving={prefSaving}
          onToggleDigestEmail={(checked) => handlePrefChange({ DigestEmail: checked })}
          onFrequencyChange={(value) => handlePrefChange({ DigestFrequency: value as "daily" | "weekly" })}
          onHourChange={(value) => handlePrefChange({ DigestHour: value })}
          onDayChange={(value) => handlePrefChange({ DigestDay: value })}
          onTimezoneChange={(value) => handlePrefChange({ DigestTimezone: value })}
          digestWebhook={pref.DigestWebhook}
          webhookURL={pref.WebhookURL}
          onToggleDigestWebhook={(checked) => handlePrefChange({ DigestWebhook: checked })}
          onWebhookURLChange={(value) => setPref((p) => ({ ...p, WebhookURL: value }))}
          onWebhookURLSave={() => updatePreference({ WebhookURL: pref.WebhookURL })}
        />

        <QuietHoursSettings
          enabled={pref.QuietHoursEnabled}
          start={pref.QuietHoursStart}
          end={pref.QuietHoursEnd}
          prefLoading={prefLoading}
          prefSaving={prefSaving}
          onEnabledChange={(checked) => handlePrefChange({ QuietHoursEnabled: checked })}
          onStartChange={(value) => handlePrefChange({ QuietHoursStart: value })}
          onEndChange={(value) => handlePrefChange({ QuietHoursEnd: value })}
        />

        <DigestStyleSettings
          digestStyle={pref.DigestStyle}
          prefLoading={prefLoading}
          prefSaving={prefSaving}
          onDigestStyleChange={(value) => handlePrefChange({ DigestStyle: value as "detailed" | "concise" })}
        />
      </SettingsPanel>

      <section id="categories" className="scroll-mt-24">
        <DigestCategoriesSettings
          excludedCategories={pref.ExcludedCategories || []}
          prefLoading={prefLoading}
          prefSaving={prefSaving}
          onExcludedCategoriesChange={(categories) => handlePrefChange({ ExcludedCategories: categories })}
        />
      </section>

      <section id="keywords" className="scroll-mt-24">
        <KeywordPreferencesSettings
          interestKeywords={pref.InterestKeywords || []}
          exclusionKeywords={pref.ExclusionKeywords || []}
          prefLoading={prefLoading}
          prefSaving={prefSaving}
          onInterestKeywordsChange={(keywords) => handlePrefChange({ InterestKeywords: keywords })}
          onExclusionKeywordsChange={(keywords) => handlePrefChange({ ExclusionKeywords: keywords })}
        />
      </section>

      <section id="analysis-rules" className="scroll-mt-24">
        <AnalysisRulesSettings />
      </section>
    </>
  );
}
