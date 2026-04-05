import { test, expect } from "@playwright/test";

test.describe("Login page", () => {
  test("shows login buttons and app branding", async ({ page }) => {
    await page.goto("/login");

    await expect(page.getByRole("heading", { name: /MailBrief/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /Continue with Google/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /Continue with GitHub/i })).toBeVisible();
  });

  test("shows demo digest preview", async ({ page }) => {
    await page.goto("/login");

    await expect(page.getByText("Your Daily Newsletter Digest")).toBeVisible();
    await expect(page.getByText("AI breakthroughs")).toBeVisible();
  });

  test("redirects unauthenticated users to login", async ({ page }) => {
    await page.goto("/inbox");

    // Should redirect to login or not-invited
    await expect(page).toHaveURL(/\/(login|not-invited)/);
  });
});
