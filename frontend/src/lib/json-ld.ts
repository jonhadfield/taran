import { APP_NAME, GITHUB_REPO_URL, SITE_DESCRIPTION, siteUrl } from "@/lib/config";

export function siteJsonLd(): Record<string, unknown> {
  const origin = siteUrl();
  return {
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
        url: origin,
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
        sameAs: [GITHUB_REPO_URL],
      },
    ],
  };
}

export function jsonLdScript(data: unknown): string {
  return JSON.stringify(data).replace(/</g, "\\u003c");
}
