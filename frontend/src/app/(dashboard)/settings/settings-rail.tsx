"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { SETTINGS_GROUPS } from "./settings-groups";

/**
 * Settings navigation, down the side rather than across the top.
 *
 * The old pill bar needed more than twice the width it had, so most of it sat
 * off-screen. A column fits every entry at once in space the page was not
 * using, and the sections of the current page sit under it as anchors.
 */
export function SettingsRail() {
  const pathname = usePathname();

  return (
    <nav aria-label="Settings" className="w-48 shrink-0">
      <ul className="sticky top-6 space-y-1 text-sm">
        {SETTINGS_GROUPS.map((group) => {
          const href = `/settings/${group.slug}`;
          const current = pathname === href;
          return (
            <li key={group.slug}>
              <Link
                href={href}
                aria-current={current ? "page" : undefined}
                className={cn(
                  "block rounded-md px-3 py-1.5 transition-colors",
                  current
                    ? "bg-muted font-medium text-foreground"
                    : "text-muted-foreground hover:bg-muted/60 hover:text-foreground",
                )}
              >
                {group.label}
              </Link>
              {current && (
                <ul className="mt-1 mb-2 space-y-0.5 border-l pl-3 ml-3">
                  {group.sections.map((section) => (
                    <li key={section.id}>
                      <a
                        href={`${href}#${section.id}`}
                        className="block rounded-md px-2 py-1 text-muted-foreground transition-colors hover:text-foreground"
                      >
                        {section.label}
                      </a>
                    </li>
                  ))}
                </ul>
              )}
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
