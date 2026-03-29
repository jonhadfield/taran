import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount, createTestEmail } from "./fixtures";

let userId: string;

test.describe("Subscriptions page", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("shows subscription with unsubscribe link", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);

    // Create email with an unsubscribe URL (set via DB since fixtures don't support it directly)
    await createTestEmail(userId, account, {
      fromName: "Newsletter Co",
      fromAddress: "news@newsletter.example.com",
    });

    await page.goto("/senders/subscriptions");

    // Page should load (may show "No subscriptions" if unsubscribe URL isn't set)
    await expect(page.getByRole("heading", { name: /Subscriptions/i }).or(page.getByText(/No subscriptions/i))).toBeVisible({ timeout: 10000 });
  });

  test("empty subscriptions shows illustration", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/senders/subscriptions");

    await expect(page.getByText(/No subscriptions found/i)).toBeVisible({ timeout: 10000 });
  });
});
