"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";
import type { ColorTheme } from "@/components/color-theme-provider";

// Swatches preview the accent each theme applies, so these are deliberately
// literal colours rather than semantic tokens — a token would render the wrong
// accent (and would make Blue and Violet identical).
const THEMES: { value: ColorTheme; label: string; swatch: string }[] = [
  { value: "neutral", label: "Neutral", swatch: "bg-zinc-800 dark:bg-zinc-300" },
  { value: "blue", label: "Blue", swatch: "bg-blue-600 dark:bg-blue-400" },
  { value: "rose", label: "Rose", swatch: "bg-rose-600 dark:bg-rose-400" },
  { value: "green", label: "Green", swatch: "bg-green-600 dark:bg-green-400" },
  { value: "violet", label: "Violet", swatch: "bg-violet-600 dark:bg-violet-400" },
  { value: "amber", label: "Amber", swatch: "bg-amber-600 dark:bg-amber-400" },
];

interface ThemeColorSettingsProps {
  colorTheme: ColorTheme;
  onColorThemeChange: (value: ColorTheme) => void;
}

export function ThemeColorSettings({
  colorTheme,
  onColorThemeChange,
}: ThemeColorSettingsProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Accent Color</CardTitle>
        <CardDescription>
          Choose a color accent for buttons, links, and highlights
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-wrap gap-3">
          {THEMES.map((theme) => (
            <button
              key={theme.value}
              type="button"
              onClick={() => onColorThemeChange(theme.value)}
              className="flex flex-col items-center gap-1.5"
              aria-label={`${theme.label} theme`}
              aria-pressed={colorTheme === theme.value}
            >
              <span
                className={cn(
                  "size-8 rounded-full transition-shadow",
                  theme.swatch,
                  colorTheme === theme.value
                    ? "ring-2 ring-ring ring-offset-2 ring-offset-background"
                    : "hover:ring-2 hover:ring-muted-foreground/30 hover:ring-offset-2 hover:ring-offset-background"
                )}
              />
              <span className="text-xs text-muted-foreground">{theme.label}</span>
            </button>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
