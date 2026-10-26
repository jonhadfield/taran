import type { MetadataRoute } from "next";

const siteUrl = (
  process.env.NEXT_PUBLIC_APP_URL || "https://mailbrief.io"
).replace(/\/$/, "");

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: "/",
      // Do not disallow `/_next/static/` — Google renders the login page with
      // those assets. Blocking them causes failed renders and soft 404s.
      disallow: [
        "/api/",
        "/inbox",
        "/settings",
        "/digests",
        "/senders",
        "/admin",
        "/onboarding",
        "/shared",
        "/not-invited",
      ],
    },
    sitemap: `${siteUrl}/sitemap.xml`,
  };
}
