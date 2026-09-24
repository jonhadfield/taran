"use client";

import { useColorTheme } from "@/components/color-theme-provider";
import { ThemeColorSettings } from "../theme-color-settings";
import { ApiKeysSettings } from "../api-keys-settings";
import { DailyLimitSettings } from "../daily-limit-settings";
import { UsageStatsCard } from "../usage-stats";
import { ExportDataButton } from "../export-data-button";
import { SignOutButton } from "../sign-out-button";
import { SettingsHeader, SettingRow, SettingsPanel } from "../settings-panel";
import { settingsGroupBySlug } from "../settings-groups";
import { usePreferences } from "../use-preferences";

const group = settingsGroupBySlug("account")!;

export default function AccountSettingsPage() {
  const { colorTheme, setColorTheme } = useColorTheme();
  const { pref, setPref, prefLoading, prefSaving, updatePreference } = usePreferences();

  return (
    <>
      <SettingsHeader title={group.label} description={group.description} />

      <SettingsPanel>
        <ThemeColorSettings colorTheme={colorTheme} onColorThemeChange={setColorTheme} />

        <DailyLimitSettings
          dailyTokenLimit={pref.DailyTokenLimit}
          prefLoading={prefLoading}
          prefSaving={prefSaving}
          onDailyTokenLimitChange={(value) => {
            setPref((p) => ({ ...p, DailyTokenLimit: value }));
            updatePreference({ DailyTokenLimit: value });
          }}
        />
      </SettingsPanel>

      <section id="api-keys" className="scroll-mt-24">
        <ApiKeysSettings />
      </section>

      <section id="usage" className="scroll-mt-24">
        <UsageStatsCard />
      </section>

      <SettingsPanel>
        <SettingRow
          id="data"
          title="Your data"
          description="Export all your emails and digests as JSON"
          control={<ExportDataButton />}
        />
        <SettingRow
          id="account"
          title="Sign out"
          description="End your session on this device"
          control={<SignOutButton />}
        />
      </SettingsPanel>
    </>
  );
}
