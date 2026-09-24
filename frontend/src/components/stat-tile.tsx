import type { LucideIcon } from "lucide-react";
import { ArrowDownRight, ArrowUpRight } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

/**
 * A single headline figure.
 *
 * The number leads, the label sits above it in muted ink, and any comparison
 * sits underneath. The icon is a quiet marker rather than a coloured block:
 * tinting each tile a different hue implied the three figures were different
 * kinds of thing, when they are all just counts.
 */
export function StatTile({
  label,
  value,
  hint,
  delta,
  icon: Icon,
  className,
}: {
  label: string;
  value: string | number;
  hint?: string;
  delta?: { value: number; suffix?: string };
  icon?: LucideIcon;
  className?: string;
}) {
  const rising = delta ? delta.value >= 0 : false;
  const DeltaIcon = rising ? ArrowUpRight : ArrowDownRight;

  return (
    <Card className={cn("gap-0 py-0", className)}>
      <CardContent className="px-5 py-4">
        <div className="flex items-start justify-between gap-3">
          <p className="text-sm text-muted-foreground">{label}</p>
          {Icon && <Icon className="size-4 shrink-0 text-muted-foreground/60" />}
        </div>
        <p className="mt-1 text-3xl font-semibold tracking-tight tabular-nums">{value}</p>
        {delta && (
          // The arrow carries the direction as well as the colour.
          <p
            className={cn(
              "mt-1.5 flex items-center gap-1 text-xs",
              rising ? "text-success" : "text-destructive",
            )}
          >
            <DeltaIcon className="size-3.5" />
            <span className="tabular-nums">
              {rising ? "+" : ""}
              {delta.value}
            </span>
            {delta.suffix && <span className="text-muted-foreground">{delta.suffix}</span>}
          </p>
        )}
        {!delta && hint && <p className="mt-1.5 text-xs text-muted-foreground">{hint}</p>}
      </CardContent>
    </Card>
  );
}
