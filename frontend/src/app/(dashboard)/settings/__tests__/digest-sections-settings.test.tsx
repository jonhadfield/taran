import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { DigestSectionsSettings } from "../digest-sections-settings";

describe("DigestSectionsSettings", () => {
  it("toggles each section", () => {
    const onHighlights = vi.fn();
    const onTopics = vi.fn();
    const onEmails = vi.fn();

    render(
      <DigestSectionsSettings
        includeHighlights
        includeTopics
        includeEmails
        prefLoading={false}
        prefSaving={false}
        onIncludeHighlightsChange={onHighlights}
        onIncludeTopicsChange={onTopics}
        onIncludeEmailsChange={onEmails}
      />,
    );

    fireEvent.click(screen.getByRole("switch", { name: /Include highlights/i }));
    fireEvent.click(screen.getByRole("switch", { name: /Include topics/i }));
    fireEvent.click(screen.getByRole("switch", { name: /Include email list/i }));

    expect(onHighlights).toHaveBeenCalledWith(false);
    expect(onTopics).toHaveBeenCalledWith(false);
    expect(onEmails).toHaveBeenCalledWith(false);
  });
});
