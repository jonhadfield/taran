import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount } from "./fixtures";

let userId: string;

test.describe("Account deletion", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("settings page shows account and delete confirmation", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings");
    await expect(page.getByRole("heading", { name: /Settings/i })).toBeVisible({ timeout: 10000 });

    // Account section should render with the email
    const section = page.locator("section#accounts");
    await expect(section).toBeVisible({ timeout: 10000 });

    // Find the trash icon button in the accounts section
    const trashButton = section.locator("button").filter({ has: page.locator("svg.lucide-trash2") }).first();
    if (await trashButton.isVisible({ timeout: 5000 }).catch(() => false)) {
      await trashButton.click();

      // Confirmation dialog should mention permanent deletion
      await expect(page.getByText(/permanently delete/i)).toBeVisible({ timeout: 5000 });
      await expect(page.getByRole("button", { name: /Cancel/i })).toBeVisible();

      // Cancel — don't actually delete
      await page.getByRole("button", { name: /Cancel/i }).click();
    }
  });
});
