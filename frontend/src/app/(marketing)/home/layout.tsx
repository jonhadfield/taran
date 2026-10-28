import type { Metadata } from "next";
import { JsonLd } from "@/components/json-ld";
import { siteJsonLd } from "@/lib/json-ld";
import { APP_NAME, SITE_DESCRIPTION, SITE_TITLE } from "@/lib/config";

export const metadata: Metadata = {
  title: { absolute: SITE_TITLE },
  description: SITE_DESCRIPTION,
  robots: { index: true, follow: true },
  alternates: { canonical: "/" },
  openGraph: {
    title: SITE_TITLE,
    description: SITE_DESCRIPTION,
    url: "/",
    siteName: APP_NAME,
    locale: "en_GB",
    type: "website",
    images: [
      {
        url: "/digest-flow.png",
        width: 1280,
        height: 720,
        alt: "Forward newsletters, digest with AI, read one brief",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: SITE_TITLE,
    description: SITE_DESCRIPTION,
    images: ["/digest-flow.png"],
  },
};

export default function HomeLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <>
      <JsonLd data={siteJsonLd()} />
      {children}
    </>
  );
}
