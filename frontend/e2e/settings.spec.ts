import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount } from "./fixtures";

let userId: string;

test.describe("Settings page", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("shows settings page with all sections", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings");

    await expect(page.getByRole("heading", { name: /Settings/i })).toBeVisible({ timeout: 10000 });

    // Check group headers (use role heading to avoid matching sidebar/card text)
    await expect(page.getByRole("heading", { name: "General" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Digest" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Organisation" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Advanced" })).toBeVisible();
  });

  test("settings nav pill bar is visible", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings");

    // Nav pills should be visible
    await expect(page.getByRole("button", { name: "Accounts" })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole("button", { name: "Delivery" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Labels" })).toBeVisible();
  });

  test("theme color picker is visible", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings");

    await expect(page.getByText("Accent Color")).toBeVisible({ timeout: 10000 });
  });
});
