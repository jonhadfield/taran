import type { Metadata } from "next";
import { cookies } from "next/headers";
import { Instrument_Sans, Newsreader, Geist_Mono } from "next/font/google";
import { APP_NAME } from "@/lib/config";
import { ThemeProvider } from "@/components/theme-provider";
import { TooltipProvider } from "@/components/ui/tooltip";
import { parseColorTheme } from "@/lib/constants";
import "./globals.css";

// Interface face. Geist is Next.js's own default, so it reads as an untouched
// install; Instrument Sans is a little narrower and sharper, which suits a tool
// for working through a queue of mail.
const interfaceSans = Instrument_Sans({
  variable: "--font-interface",
  subsets: ["latin"],
  display: "swap",
});

// Reading face, used only for what the product produces: AI summaries and
// digest prose. Newsreader is drawn for editorial reading on screen, which is
// exactly the job — the interface is an instrument, the digest is something you
// read.
const readingSerif = Newsreader({
  variable: "--font-newsreader",
  subsets: ["latin"],
  display: "swap",
});

// Kept for addresses, tokens and other data that should not be mistaken for prose.
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
  const colorTheme = parseColorTheme(cookieStore.get("color-theme")?.value);

  return (
    <html lang="en" suppressHydrationWarning {...(colorTheme !== "brand" ? { "data-theme": colorTheme } : {})}>
      <body
        className={`${interfaceSans.variable} ${readingSerif.variable} ${geistMono.variable} antialiased`}
      >
        <ThemeProvider>
          <TooltipProvider delayDuration={300}>{children}</TooltipProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
