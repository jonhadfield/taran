import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { Header } from "@/components/header";
import { Sidebar } from "@/components/sidebar";
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
        {/* Fixed full-height sidebar - hidden on mobile */}
        <aside className="hidden lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col">
          <Sidebar isAdmin={admin} />
        </aside>

        {/* Main content area offset by sidebar width */}
        <div className="lg:pl-64">
          <Header isAdmin={admin} />
          <main className="@container p-4 lg:p-6">{children}</main>
        </div>
        <CommandPalette isAdmin={admin} />
        <KeyboardHelp />
        <Toaster />
      </div>
    </ColorThemeProvider>
  );
}
