import Link from "next/link";
import { ChevronRight } from "lucide-react";
import { SETTINGS_GROUPS } from "./settings-groups";

export default function SettingsHubPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Settings</h1>

      <nav aria-label="Settings groups" className="divide-y rounded-xl border bg-card">
        {SETTINGS_GROUPS.map((group) => (
          <Link
            key={group.slug}
            href={`/settings/${group.slug}`}
            className="flex items-center justify-between gap-4 px-5 py-4 transition-colors first:rounded-t-xl last:rounded-b-xl hover:bg-muted/50"
          >
            <div className="min-w-0 space-y-1">
              <h2 className="text-sm font-medium">{group.label}</h2>
              <p className="text-sm text-muted-foreground">{group.description}</p>
            </div>
            <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
          </Link>
        ))}
      </nav>
    </div>
  );
}
