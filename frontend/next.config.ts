import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Next 16.2+ writes AGENTS.md and CLAUDE.md into the project root on `next
  // dev` when it detects an AI coding agent, and re-creates them if deleted.
  // Disabled here so the framework does not generate tracked files as a side
  // effect of running the dev server; agent configuration for this repo is
  // managed deliberately rather than by the build tool.
  agentRules: false,
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "lh3.googleusercontent.com",
      },
      {
        protocol: "https",
        hostname: "avatars.githubusercontent.com",
      },
    ],
  },
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          { key: "X-Frame-Options", value: "DENY" },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          { key: "X-DNS-Prefetch-Control", value: "on" },
          {
            key: "Permissions-Policy",
            value: "camera=(), microphone=(), geolocation=()",
          },
        ],
      },
    ];
  },
};

export default nextConfig;
