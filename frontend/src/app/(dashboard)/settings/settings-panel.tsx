import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

/**
 * One surface holding several settings, separated by hairlines.
 *
 * Giving every individual setting its own card made the page long and flat:
 * a one-line toggle carried the same weight as a whole guide. Cards are kept
 * for collections you add to, such as labels or rules.
 */
export function SettingsPanel({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("divide-y rounded-xl border bg-card", className)}>
      {children}
    </div>
  );
}

/**
 * A single setting: what it is on the left, its control on the right.
 * `layout="stacked"` puts the control underneath, for wide controls such as
 * keyword inputs or a row of swatches.
 */
export function SettingRow({
  id,
  title,
  description,
  control,
  children,
  layout = "inline",
}: {
  id?: string;
  title: string;
  description?: ReactNode;
  control?: ReactNode;
  children?: ReactNode;
  layout?: "inline" | "stacked";
}) {
  return (
    <div id={id} className="scroll-mt-24 px-5 py-4">
      <div
        className={cn(
          "gap-4",
          layout === "inline" ? "flex items-start justify-between" : "space-y-3",
        )}
      >
        <div className="min-w-0 space-y-1">
          <h3 className="text-sm font-medium leading-none">{title}</h3>
          {description && (
            <p className="text-sm text-muted-foreground">{description}</p>
          )}
        </div>
        {control && <div className="shrink-0">{control}</div>}
        {layout === "stacked" && children}
      </div>
      {layout === "inline" && children && <div className="mt-4">{children}</div>}
    </div>
  );
}

/** Heading for a page of settings. */
export function SettingsHeader({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="space-y-1">
      <h1 className="text-2xl font-bold tracking-tight">{title}</h1>
      <p className="text-sm text-muted-foreground">{description}</p>
    </div>
  );
}
