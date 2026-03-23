import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount } from "./fixtures";

let userId: string;

test.describe("Dashboard", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("shows dashboard for users with an account", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/");

    await expect(page.getByRole("heading", { name: /Dashboard/i })).toBeVisible({ timeout: 10000 });
  });

  test("shows empty state when no emails", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/");

    await expect(page.getByText(/No emails yet/i)).toBeVisible({ timeout: 10000 });
  });

  test("sidebar navigation works", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/");
    await expect(page.getByRole("heading", { name: /Dashboard/i })).toBeVisible({ timeout: 10000 });

    // Navigate to inbox via sidebar
    await page.getByRole("link", { name: /Inbox/i }).first().click();
    await expect(page).toHaveURL(/\/inbox/);

    // Navigate to digests
    await page.getByRole("link", { name: /Digests/i }).first().click();
    await expect(page).toHaveURL(/\/digests/);

    // Navigate to senders
    await page.getByRole("link", { name: /Senders/i }).first().click();
    await expect(page).toHaveURL(/\/senders/);

    // Navigate to settings
    await page.getByRole("link", { name: /Settings/i }).first().click();
    await expect(page).toHaveURL(/\/settings/);
  });
});
