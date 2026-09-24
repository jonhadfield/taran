"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronLeft } from "lucide-react";
import { SettingsRail } from "./settings-rail";

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isHub = pathname === "/settings";

  // The hub is the menu, so it needs no navigation beside it.
  if (isHub) {
    return <div className="max-w-2xl">{children}</div>;
  }

  return (
    <div className="flex gap-10">
      <div className="hidden md:block">
        <SettingsRail />
      </div>
      <div className="min-w-0 max-w-2xl flex-1 space-y-6">
        <Link
          href="/settings"
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground md:hidden"
        >
          <ChevronLeft className="size-4" />
          Settings
        </Link>
        {children}
      </div>
    </div>
  );
}
