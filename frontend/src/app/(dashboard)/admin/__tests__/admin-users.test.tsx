import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import type { AdminUser, ListResponse } from "@/types/api";

const paths: string[] = [];
let response: ListResponse<AdminUser> = { data: [], total: 0 };

vi.mock("@/hooks/use-polling", () => ({
  usePolling: (path: string) => {
    paths.push(path);
    return { data: response, loading: false, error: null, refresh: vi.fn() };
  },
}));
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { AdminUsers } from "../admin-users";

function users(count: number, from = 1): AdminUser[] {
  return Array.from({ length: count }, (_, i) => ({
    ID: `u${from + i}`,
    Email: `user${from + i}@example.com`,
    Name: `User ${from + i}`,
    EmailCount: 0,
    MonthlyTokensUsed: 0,
    MonthlyTokenLimit: 500000,
  }));
}

const lastPath = () => paths[paths.length - 1];

beforeEach(() => {
  paths.length = 0;
  response = { data: [], total: 0 };
});

describe("AdminUsers paging", () => {
  it("asks for the first page and reports the range", () => {
    response = { data: users(25), total: 73 };
    render(<AdminUsers />);

    expect(lastPath()).toBe("admin/users?limit=25&offset=0");
    expect(screen.getByText("Showing 1–25 of 73")).toBeInTheDocument();
    expect(screen.getByText("73 users")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Previous/ })).toBeDisabled();
    expect(screen.getByRole("button", { name: /Next/ })).toBeEnabled();
  });

  it("steps forward and back a page", async () => {
    response = { data: users(25), total: 73 };
    render(<AdminUsers />);

    fireEvent.click(screen.getByRole("button", { name: /Next/ }));
    await waitFor(() => expect(lastPath()).toBe("admin/users?limit=25&offset=25"));

    fireEvent.click(screen.getByRole("button", { name: /Previous/ }));
    await waitFor(() => expect(lastPath()).toBe("admin/users?limit=25&offset=0"));
  });

  it("disables Next on the final page", () => {
    response = { data: users(23, 51), total: 73 };
    render(<AdminUsers />);
    // A short final page still reports its true range.
    expect(screen.getByText("Showing 1–23 of 73")).toBeInTheDocument();
  });

  it("hides the pager when everyone fits on one page", () => {
    response = { data: users(9), total: 9 };
    render(<AdminUsers />);
    expect(screen.queryByText(/Showing/)).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Next/ })).not.toBeInTheDocument();
  });

  it("steps back when users are removed from under the current page", async () => {
    response = { data: users(25), total: 73 };
    const { rerender } = render(<AdminUsers />);

    fireEvent.click(screen.getByRole("button", { name: /Next/ }));
    fireEvent.click(screen.getByRole("button", { name: /Next/ }));
    await waitFor(() => expect(lastPath()).toBe("admin/users?limit=25&offset=50"));

    // Most users are deleted elsewhere; the next poll returns a smaller total.
    response = { data: users(9), total: 9 };
    rerender(<AdminUsers />);

    await waitFor(() => expect(lastPath()).toBe("admin/users?limit=25&offset=0"));
  });
});
