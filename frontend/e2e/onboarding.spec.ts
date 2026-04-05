import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";

let userId: string;

test.describe("Onboarding flow", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("redirects new users to onboarding", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;

    await page.goto("/");

    // New users with no email account should land on onboarding
    await expect(page).toHaveURL(/\/onboarding/);
  });

  test("shows username form on step 2", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;

    await page.goto("/onboarding");

    // Click through step 1
    await page.getByRole("button", { name: /get started/i, exact: false }).first().click();

    // Step 2 should show a username input
    await expect(page.getByPlaceholder(/username/i)).toBeVisible({ timeout: 5000 });
  });
});
