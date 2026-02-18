import { describe, it, expect } from "vitest";
import { buildQueryString } from "../inbox-list";

describe("buildQueryString", () => {
  it("returns default params for 'all' filter", () => {
    expect(buildQueryString("all", "", "", 50)).toBe("limit=50");
  });

  it("adds is_read=false for 'unread' filter", () => {
    expect(buildQueryString("unread", "", "", 50)).toBe("limit=50&is_read=false");
  });

  it("adds is_starred=true for 'starred' filter", () => {
    expect(buildQueryString("starred", "", "", 50)).toBe("limit=50&is_starred=true");
  });

  it("adds is_archived=true for 'archived' filter", () => {
    expect(buildQueryString("archived", "", "", 50)).toBe("limit=50&is_archived=true");
  });

  it("adds search param when provided", () => {
    expect(buildQueryString("all", "hello world", "", 50)).toBe(
      "limit=50&search=hello%20world",
    );
  });

  it("adds topic param when provided", () => {
    expect(buildQueryString("all", "", "Tech", 50)).toBe("limit=50&topic=Tech");
  });

  it("combines filter, search, and topic", () => {
    expect(buildQueryString("unread", "test", "News", 25)).toBe(
      "limit=25&is_read=false&search=test&topic=News",
    );
  });

  it("uses custom limit", () => {
    expect(buildQueryString("all", "", "", 100)).toBe("limit=100");
  });
});
