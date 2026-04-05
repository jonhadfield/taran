import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount, createTestEmail, createTestDigest } from "./fixtures";

let userId: string;

test.describe("Digests", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("shows digest in list", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account);
    await createTestDigest(userId, [email], {
      title: "Weekly Digest — March 22",
    });

    await page.goto("/digests");

    await expect(page.getByText("Weekly Digest — March 22")).toBeVisible({ timeout: 10000 });
  });

  test("clicking digest shows summary and highlights", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account, {
      subject: "AI News This Week",
      fromName: "AI Digest",
    });
    await createTestDigest(userId, [email], {
      title: "Weekly Digest — AI Focus",
      summary: "This week featured major AI announcements and new model releases.",
    });

    await page.goto("/digests");

    await page.getByText("Weekly Digest — AI Focus").click();

    // Summary card
    await expect(page.getByText("This week featured major AI announcements")).toBeVisible({ timeout: 10000 });

    // Highlights
    await expect(page.getByText("AI developments continue")).toBeVisible({ timeout: 10000 });

    // Topics section should render
    await expect(page.getByText("Top Topics")).toBeVisible({ timeout: 5000 });

    // Included emails
    await expect(page.getByText("Included Emails")).toBeVisible({ timeout: 5000 });
    await expect(page.getByText("AI News This Week")).toBeVisible({ timeout: 5000 });
  });

  test("empty digests page shows illustration", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/digests");

    await expect(page.getByText("No digests yet")).toBeVisible({ timeout: 10000 });
  });
});
