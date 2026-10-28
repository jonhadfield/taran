import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import HomePage from "../page";

describe("HomePage", () => {
  it("explains the product and points at sign-in", () => {
    render(<HomePage />);

    expect(
      screen.getByRole("heading", {
        level: 1,
        name: /Forward your newsletters\. Read one brief\./i,
      }),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Get started/i })).toHaveAttribute(
      "href",
      "/login",
    );
    expect(screen.getByRole("link", { name: /Source on GitHub/i })).toHaveAttribute(
      "href",
      "https://github.com/jonhadfield/taran",
    );
    expect(screen.getByRole("heading", { name: /How it works/i })).toBeInTheDocument();
    expect(screen.getByText("Your Daily Newsletter Digest")).toBeInTheDocument();
  });
});
