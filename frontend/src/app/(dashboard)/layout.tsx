import { cookies, headers } from "next/headers";
import { redirect } from "next/navigation";
import { DashboardShell } from "@/components/dashboard-shell";
import { TitleUpdater } from "@/components/title-updater";
import { Toaster } from "@/components/ui/sonner";
import { ColorThemeProvider } from "@/components/color-theme-provider";
import { isAdmin } from "@/lib/admin";
import { auth } from "@/lib/auth";
import { checkAccess } from "@/lib/access";
import { AccessUnavailable } from "@/components/access-unavailable";
import { CommandPalette } from "@/components/command-palette";
import { KeyboardHelp } from "@/components/keyboard-help";
import { parseColorTheme } from "@/lib/constants";

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const cookieStore = await cookies();
  const colorTheme = parseColorTheme(cookieStore.get("color-theme")?.value);

  // Validate session first — redirect to login if expired or missing
  const session = await auth.api.getSession({ headers: await headers() });
  if (!session) {
    redirect("/login");
  }

  const admin = await isAdmin();

  if (!admin) {
    // redirect() throws, so it must run outside the check's own error handling.
    const access = await checkAccess();
    if (access === "unauthenticated") {
      redirect("/login");
    }
    if (access === "denied") {
      redirect("/not-invited");
    }
    if (access === "unavailable") {
      return <AccessUnavailable />;
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
