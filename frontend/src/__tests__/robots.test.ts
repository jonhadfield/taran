// @vitest-environment node
import { describe, it, expect } from "vitest";
import robots from "../app/robots";
import sitemap from "../app/sitemap";

describe("robots.txt", () => {
  it("does not block Next.js static assets needed to render public pages", () => {
    const { rules } = robots();
    const rule = Array.isArray(rules) ? rules[0] : rules;
    const disallow = [rule?.disallow].flat().filter(Boolean);
    expect(disallow.join(" ")).not.toMatch(/_next/);
  });

  it("keeps authenticated app routes out of the crawl", () => {
    const { rules } = robots();
    const rule = Array.isArray(rules) ? rules[0] : rules;
    const disallow = [rule?.disallow].flat();
    expect(disallow).toEqual(
      expect.arrayContaining(["/api/", "/inbox", "/shared"]),
    );
  });

  it("points at the sitemap", () => {
    expect(robots().sitemap).toMatch(/\/sitemap\.xml$/);
  });
});

describe("sitemap", () => {
  it("lists the public login page", () => {
    const entries = sitemap();
    expect(entries.map((entry) => entry.url)).toEqual(
      expect.arrayContaining([expect.stringMatching(/\/login$/)]),
    );
  });
});
