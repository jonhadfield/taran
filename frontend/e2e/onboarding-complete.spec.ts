import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";

let userId: string;

test.describe("Onboarding complete flow", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("full onboarding: step 1 → step 2 → step 3 → dashboard", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;

    await page.goto("/");

    // Should redirect to onboarding
    await expect(page).toHaveURL(/\/onboarding/, { timeout: 10000 });

    // Step 1: Get Started
    await expect(page.getByText("Welcome to")).toBeVisible({ timeout: 10000 });
    await page.getByRole("button", { name: /Get Started/i }).first().click();

    // Step 2: Username form
    await expect(page.getByPlaceholder(/username/i)).toBeVisible({ timeout: 5000 });

    // Enter a unique username
    const username = `test-${Date.now()}`;
    await page.getByPlaceholder(/username/i).fill(username);

    // Wait for availability check
    await page.waitForTimeout(600); // debounce + API call

    // Submit
    await page.getByRole("button", { name: /Create Inbox/i }).click();

    // Step 3: Success — should show the email address
    await expect(page.getByText("Your inbox is ready")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(`${username}@`)).toBeVisible();

    // Click to go to dashboard
    await page.getByRole("button", { name: /Go to Dashboard/i }).click();

    // Should land on inbox
    await expect(page).toHaveURL(/\/inbox/, { timeout: 10000 });
  });
});
