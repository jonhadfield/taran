import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount, createTestEmail } from "./fixtures";

let userId: string;

test.describe("Inbox bulk actions", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("select all shows bulk action bar", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    await createTestEmail(userId, account, { subject: "Bulk Email 1" });
    await createTestEmail(userId, account, { subject: "Bulk Email 2" });

    await page.goto("/inbox");
    await expect(page.getByText("Bulk Email 1")).toBeVisible({ timeout: 10000 });

    // Select all via the header checkbox
    await page.getByLabel("Select all emails").check();

    // Bulk action bar should appear with selection count
    await expect(page.getByText(/2 of 2 selected/i)).toBeVisible({ timeout: 5000 });
  });

  test("individual email checkbox toggles selection", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    await createTestEmail(userId, account, { subject: "Checkbox Test" });

    await page.goto("/inbox");
    await expect(page.getByText("Checkbox Test")).toBeVisible({ timeout: 10000 });

    // Check the individual email checkbox
    const checkbox = page.getByLabel(/Select Checkbox Test/i);
    await checkbox.check();

    // Should show selection count
    await expect(page.getByText(/1 of 1 selected/i)).toBeVisible({ timeout: 5000 });
  });
});
