import type { MetadataRoute } from "next";
import { siteUrl } from "@/lib/config";

export default function sitemap(): MetadataRoute.Sitemap {
  return [
    {
      url: `${siteUrl()}/login`,
      changeFrequency: "monthly",
      priority: 1,
    },
  ];
}
