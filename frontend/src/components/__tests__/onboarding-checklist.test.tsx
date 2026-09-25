import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { OnboardingChecklist } from "../onboarding-checklist";

const DISMISS_KEY = "onboarding-checklist-dismissed";

/** A brand new account: inbox created, nothing else done yet. */
function freshProps() {
  return {
    emailAddress: "someone@example.com",
    hasEmails: false,
    hasDigest: false,
    hasConfiguredPreferences: false as boolean | null,
  };
}

/** An account that finished onboarding. */
function completeProps() {
  return {
    emailAddress: "someone@example.com",
    hasEmails: true,
    hasDigest: true,
    hasConfiguredPreferences: true as boolean | null,
  };
}

describe("OnboardingChecklist", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("guides a new account through the remaining steps", () => {
    render(<OnboardingChecklist {...freshProps()} />);
    expect(screen.getByText("Getting started")).toBeInTheDocument();
    expect(screen.getByText("1 of 4 complete")).toBeInTheDocument();
  });

  it("renders nothing while the preferences lookup is still in flight", () => {
    // Rendering here would show a stale count and then rip it away, which is
    // what made the checklist flash past on every dashboard load.
    const { container } = render(
      <OnboardingChecklist {...completeProps()} hasConfiguredPreferences={null} />
    );
    expect(container).toBeEmptyDOMElement();
  });

  it("renders nothing once every step is done", () => {
    const { container } = render(<OnboardingChecklist {...completeProps()} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("stays hidden for a finished account even on a browser that never dismissed it", () => {
    // The dismissal flag lives in localStorage, so it is absent on a new
    // browser or after clearing site data. Completion alone must be enough.
    expect(localStorage.getItem(DISMISS_KEY)).toBeNull();
    const { container } = render(<OnboardingChecklist {...completeProps()} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("respects a dismissal while onboarding is still incomplete", () => {
    localStorage.setItem(DISMISS_KEY, "1");
    const { container } = render(<OnboardingChecklist {...freshProps()} />);
    expect(container).toBeEmptyDOMElement();
  });
});
