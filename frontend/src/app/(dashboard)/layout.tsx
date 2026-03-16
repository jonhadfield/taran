import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { Header } from "@/components/header";
import { Sidebar } from "@/components/sidebar";
import { TitleUpdater } from "@/components/title-updater";
import { Toaster } from "@/components/ui/sonner";
import { ColorThemeProvider } from "@/components/color-theme-provider";
import type { ColorTheme } from "@/components/color-theme-provider";
import { isAdmin } from "@/lib/admin";
import { serverFetch } from "@/lib/server-api";
import type { AccessCheck } from "@/types/api";

const VALID_COLOR_THEMES = new Set(["neutral", "blue", "rose", "green", "violet", "amber"]);

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const cookieStore = await cookies();
  const rawTheme = cookieStore.get("color-theme")?.value;
  const colorTheme = (rawTheme && VALID_COLOR_THEMES.has(rawTheme) ? rawTheme : "neutral") as ColorTheme;

  const admin = await isAdmin();

  if (!admin) {
    try {
      const access = await serverFetch<AccessCheck>("access");
      if (!access.hasAccess) {
        redirect("/not-invited");
      }
    } catch {
      // If the access check fails (e.g. backend down), deny access
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
        <Toaster />
      </div>
    </ColorThemeProvider>
  );
}
