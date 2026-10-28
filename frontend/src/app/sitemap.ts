import type { MetadataRoute } from "next";
import { siteUrl } from "@/lib/config";

export default function sitemap(): MetadataRoute.Sitemap {
  const origin = siteUrl();
  return [
    {
      url: origin,
      changeFrequency: "monthly",
      priority: 1,
    },
    {
      url: `${origin}/login`,
      changeFrequency: "monthly",
      priority: 0.5,
    },
  ];
}
