import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { DashboardShell } from "@/components/dashboard-shell";
import { TitleUpdater } from "@/components/title-updater";
import { Toaster } from "@/components/ui/sonner";
import { ColorThemeProvider } from "@/components/color-theme-provider";
import { isAdmin } from "@/lib/admin";
import { serverFetch } from "@/lib/server-api";
import { CommandPalette } from "@/components/command-palette";
import { KeyboardHelp } from "@/components/keyboard-help";
import { parseColorTheme } from "@/lib/constants";
import type { AccessCheck } from "@/types/api";

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const cookieStore = await cookies();
  const colorTheme = parseColorTheme(cookieStore.get("color-theme")?.value);

  const admin = await isAdmin();

  if (!admin) {
    try {
      const access = await serverFetch<AccessCheck>("access");
      if (!access.hasAccess) {
        redirect("/not-invited");
      }
    } catch (err) {
      // Distinguish auth failure (expired/rotated token) from access denial.
      // Auth failures should go to login, not the "not invited" page.
      const msg = err instanceof Error ? err.message : "";
      if (msg.includes("Not authenticated") || msg.includes("401")) {
        redirect("/login");
      }
      redirect("/not-invited");
    }
  }

  return (
    <ColorThemeProvider initialTheme={colorTheme}>
      <div className="min-h-screen">
        <TitleUpdater />
        <DashboardShell isAdmin={admin}>
          {children}
        </DashboardShell>
        <CommandPalette isAdmin={admin} />
        <KeyboardHelp />
        <Toaster />
      </div>
    </ColorThemeProvider>
  );
}
