import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";

vi.mock("@/lib/auth-client", () => ({ authClient: { signIn: { social: vi.fn() } } }));
vi.mock("@/components/public-shell", () => ({
  PublicShell: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

import LoginPage from "../page";

afterEach(() => {
  vi.unstubAllGlobals();
});

function stubStatus(response: Promise<unknown>) {
  vi.stubGlobal("fetch", vi.fn(() => response));
}

describe("LoginPage access note", () => {
  it("invites sign-ups while open registration is on", async () => {
    stubStatus(Promise.resolve({ json: () => Promise.resolve({ waitlistEnabled: false, openRegistration: true }) }));
    render(<LoginPage />);
    expect(await screen.findByText(/Open for sign-ups/)).toBeInTheDocument();
    expect(screen.queryByText(/Invite-only/)).not.toBeInTheDocument();
  });

  it("says invite-only otherwise", async () => {
    stubStatus(Promise.resolve({ json: () => Promise.resolve({ waitlistEnabled: true, openRegistration: false }) }));
    render(<LoginPage />);
    expect(await screen.findByText(/Invite-only/)).toBeInTheDocument();
  });

  it("falls back to invite-only when the status can't be loaded", async () => {
    stubStatus(Promise.reject(new Error("network")));
    render(<LoginPage />);
    expect(await screen.findByText(/Invite-only/)).toBeInTheDocument();
  });
});
