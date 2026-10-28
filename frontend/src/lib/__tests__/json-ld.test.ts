import { describe, expect, it } from "vitest";
import { jsonLdScript, siteJsonLd } from "../json-ld";

describe("siteJsonLd", () => {
  it("describes the public site, not the login path", () => {
    const graph = siteJsonLd();
    expect(JSON.stringify(graph)).toContain("https://mailbrief.io");
    expect(JSON.stringify(graph)).not.toContain("/login");
    expect(jsonLdScript({ html: "<script>" })).toContain("\\u003cscript>");
  });
});
