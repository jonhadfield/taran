import type { Metadata } from "next";
import { APP_NAME, SITE_DESCRIPTION, SITE_TITLE, siteUrl } from "@/lib/config";

const origin = siteUrl();
const loginUrl = `${origin}/login`;

export const metadata: Metadata = {
  title: SITE_TITLE,
  description: SITE_DESCRIPTION,
  robots: { index: true, follow: true },
  alternates: { canonical: "/login" },
  openGraph: {
    title: SITE_TITLE,
    description: SITE_DESCRIPTION,
    url: "/login",
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

const jsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "WebSite",
      name: APP_NAME,
      url: origin,
    },
    {
      "@type": "SoftwareApplication",
      name: APP_NAME,
      url: loginUrl,
      applicationCategory: "ProductivityApplication",
      operatingSystem: "Web",
      description: SITE_DESCRIPTION,
      offers: {
        "@type": "Offer",
        price: "0",
        priceCurrency: "USD",
      },
    },
    {
      "@type": "Organization",
      name: APP_NAME,
      url: origin,
      logo: `${origin}/logo-512.png`,
      sameAs: ["https://github.com/jonhadfield/taran"],
    },
  ],
};

export default function LoginLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(jsonLd).replace(/</g, "\\u003c"),
        }}
      />
      {children}
    </>
  );
}
