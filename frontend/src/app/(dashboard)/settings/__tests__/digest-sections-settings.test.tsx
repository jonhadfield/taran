import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { DigestSectionsSettings } from "../digest-sections-settings";

describe("DigestSectionsSettings", () => {
  it("toggles each section and reorders", () => {
    const onHighlights = vi.fn();
    const onTopics = vi.fn();
    const onEmails = vi.fn();
    const onOrder = vi.fn();
    const onDensity = vi.fn();

    render(
      <DigestSectionsSettings
        includeHighlights
        includeTopics
        includeEmails
        sectionOrder={["highlights", "topics", "emails"]}
        density="comfortable"
        prefLoading={false}
        prefSaving={false}
        onIncludeHighlightsChange={onHighlights}
        onIncludeTopicsChange={onTopics}
        onIncludeEmailsChange={onEmails}
        onSectionOrderChange={onOrder}
        onDensityChange={onDensity}
      />,
    );

    fireEvent.click(screen.getByRole("switch", { name: /Include highlights/i }));
    fireEvent.click(screen.getByRole("switch", { name: /Include topics/i }));
    fireEvent.click(screen.getByRole("switch", { name: /Include email list/i }));

    expect(onHighlights).toHaveBeenCalledWith(false);
    expect(onTopics).toHaveBeenCalledWith(false);
    expect(onEmails).toHaveBeenCalledWith(false);

    fireEvent.click(screen.getByRole("button", { name: /Move Highlights down/i }));
    expect(onOrder).toHaveBeenCalledWith(["topics", "highlights", "emails"]);

    fireEvent.click(screen.getByRole("button", { name: /^Compact$/i }));
    expect(onDensity).toHaveBeenCalledWith("compact");
  });
});
