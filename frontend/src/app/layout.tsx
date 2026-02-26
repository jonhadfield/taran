import type { Metadata } from "next";
import { cookies } from "next/headers";
import { Geist, Geist_Mono } from "next/font/google";
import { APP_NAME } from "@/lib/config";
import { ThemeProvider } from "@/components/theme-provider";
import type { ColorTheme } from "@/components/color-theme-provider";
import "./globals.css";

const VALID_COLOR_THEMES = new Set(["neutral", "blue", "rose", "green", "violet", "amber"]);

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: `${APP_NAME} - Email Digest Dashboard`,
  description: "AI-powered email digest dashboard for your newsletters",
  icons: {
    icon: [
      { url: "/favicon.ico", sizes: "48x48" },
      { url: "/icon.svg", type: "image/svg+xml" },
    ],
    apple: "/apple-icon.png",
  },
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const cookieStore = await cookies();
  const rawTheme = cookieStore.get("color-theme")?.value;
  const colorTheme = (rawTheme && VALID_COLOR_THEMES.has(rawTheme) ? rawTheme : "neutral") as ColorTheme;

  return (
    <html lang="en" suppressHydrationWarning {...(colorTheme !== "neutral" ? { "data-theme": colorTheme } : {})}>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <ThemeProvider>{children}</ThemeProvider>
      </body>
    </html>
  );
}
