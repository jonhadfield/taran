import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount, createTestEmail } from "./fixtures";

let userId: string;

test.describe("Senders", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("shows sender after receiving an email", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    await createTestEmail(userId, account, {
      fromName: "Tech Weekly",
      fromAddress: "tech@weekly.example.com",
    });

    await page.goto("/senders");

    await expect(page.getByText("Tech Weekly")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText("tech@weekly.example.com")).toBeVisible();
  });

  test("empty senders page shows illustration", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/senders");

    await expect(page.getByText("No senders yet")).toBeVisible({ timeout: 10000 });
  });
});
