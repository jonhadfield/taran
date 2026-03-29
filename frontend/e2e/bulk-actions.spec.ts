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

  test("select all and bulk archive", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    await createTestEmail(userId, account, { subject: "Bulk Email 1" });
    await createTestEmail(userId, account, { subject: "Bulk Email 2" });

    await page.goto("/inbox");
    await expect(page.getByText("Bulk Email 1")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText("Bulk Email 2")).toBeVisible();

    // Select all via checkbox
    await page.getByLabel("Select all emails").check();

    // Bulk action bar should appear with count
    await expect(page.getByText(/2 of 2 selected/i)).toBeVisible({ timeout: 5000 });

    // Archive button should be visible in bulk bar
    await expect(page.getByRole("button", { name: /Archive/i })).toBeVisible();
  });

  test("select individual email with checkbox", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    await createTestEmail(userId, account, { subject: "Select Me Email" });

    await page.goto("/inbox");
    await expect(page.getByText("Select Me Email")).toBeVisible({ timeout: 10000 });

    // Check the individual email checkbox
    const checkbox = page.getByLabel(/Select Select Me Email/i);
    await checkbox.check();

    // Should show 1 selected
    await expect(page.getByText(/1 of 1 selected/i)).toBeVisible({ timeout: 5000 });
  });
});
