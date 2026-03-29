import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount, createTestEmail } from "./fixtures";

let userId: string;

test.describe("Account deletion", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("delete inbox shows confirmation with warning", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    await createTestEmail(userId, account, { subject: "Will Be Deleted" });

    await page.goto("/settings");
    await expect(page.getByRole("heading", { name: /Settings/i })).toBeVisible({ timeout: 10000 });

    // The account email should be visible
    await expect(page.getByText(account.emailAddress)).toBeVisible({ timeout: 10000 });

    // Click the trash icon button in the accounts section
    const accountSection = page.locator("#accounts");
    await accountSection.getByRole("button").filter({ has: page.locator("svg") }).first().click();

    // Confirmation dialog should appear
    await expect(page.getByText(/permanently delete/i)).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/cannot be undone/i)).toBeVisible();
    await expect(page.getByRole("button", { name: /Cancel/i })).toBeVisible();
  });

  test("delete inbox and verify redirect to onboarding", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings");
    await expect(page.getByRole("heading", { name: /Settings/i })).toBeVisible({ timeout: 10000 });

    // Click the trash icon
    const accountSection = page.locator("#accounts");
    await accountSection.getByRole("button").filter({ has: page.locator("svg") }).first().click();

    // Confirm deletion in the dialog
    await expect(page.getByText(/permanently delete/i)).toBeVisible({ timeout: 5000 });
    await page.getByRole("button", { name: /Delete/i }).click();

    // Should show success toast
    await expect(page.getByText("Inbox deleted")).toBeVisible({ timeout: 10000 });

    // Navigate to dashboard — should redirect to onboarding
    await page.goto("/");
    await expect(page).toHaveURL(/\/onboarding/, { timeout: 10000 });
  });
});
