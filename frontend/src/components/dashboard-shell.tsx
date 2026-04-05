"use client";

import { Sidebar } from "@/components/sidebar";
import { Header } from "@/components/header";
import { useSidebarCollapsed } from "@/hooks/use-sidebar-collapsed";
import { cn } from "@/lib/utils";

export function DashboardShell({
  isAdmin,
  children,
}: {
  isAdmin?: boolean;
  children: React.ReactNode;
}) {
  const { collapsed, toggle } = useSidebarCollapsed();

  return (
    <>
      {/* Fixed full-height sidebar - hidden on mobile */}
      <aside
        className={cn(
          "hidden lg:fixed lg:inset-y-0 lg:flex lg:flex-col transition-[width] duration-200",
          collapsed ? "lg:w-16" : "lg:w-64"
        )}
      >
        <Sidebar isAdmin={isAdmin} collapsed={collapsed} onToggle={toggle} />
      </aside>

      {/* Main content area offset by sidebar width */}
      <div
        className={cn(
          "transition-[padding-left] duration-200",
          collapsed ? "lg:pl-16" : "lg:pl-64"
        )}
      >
        <Header isAdmin={isAdmin} />
        <main className="@container p-4 lg:p-6">{children}</main>
      </div>
    </>
  );
}
