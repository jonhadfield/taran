"use client";

import { createContext, useCallback, useContext, useState } from "react";
import { apiPatch } from "@/lib/api";

export type ColorTheme = "neutral" | "blue" | "rose" | "green" | "violet" | "amber";

interface ColorThemeContextValue {
  colorTheme: ColorTheme;
  setColorTheme: (theme: ColorTheme) => void;
}

const ColorThemeContext = createContext<ColorThemeContextValue | undefined>(undefined);

export function ColorThemeProvider({
  initialTheme,
  children,
}: {
  initialTheme: ColorTheme;
  children: React.ReactNode;
}) {
  const [colorTheme, setColorThemeState] = useState<ColorTheme>(initialTheme);

  const setColorTheme = useCallback((theme: ColorTheme) => {
    setColorThemeState(theme);

    // Update data-theme on <html>
    if (theme === "neutral") {
      document.documentElement.removeAttribute("data-theme");
    } else {
      document.documentElement.setAttribute("data-theme", theme);
    }

    // Set cookie (365 days)
    document.cookie = `color-theme=${theme};path=/;max-age=${365 * 24 * 60 * 60};SameSite=Lax`;

    // Fire-and-forget PATCH to persist server-side
    apiPatch("preferences", { ColorTheme: theme }).catch(() => {});
  }, []);

  return (
    <ColorThemeContext.Provider value={{ colorTheme, setColorTheme }}>
      {children}
    </ColorThemeContext.Provider>
  );
}

export function useColorTheme() {
  const ctx = useContext(ColorThemeContext);
  if (!ctx) {
    throw new Error("useColorTheme must be used within a ColorThemeProvider");
  }
  return ctx;
}
