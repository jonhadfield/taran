import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount } from "./fixtures";

let userId: string;

test.describe("Label management", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("create a new label", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings#labels");
    await expect(page.getByRole("heading", { name: /Settings/i })).toBeVisible({ timeout: 10000 });

    // Click create label button
    const createButton = page.getByRole("button", { name: /Create/i });
    await createButton.scrollIntoViewIfNeeded();
    await createButton.click();

    // Fill in label name
    const nameInput = page.getByPlaceholder(/Label name/i);
    await expect(nameInput).toBeVisible({ timeout: 5000 });
    await nameInput.fill("Important");

    // Save
    await page.getByRole("button", { name: /Save/i }).click();

    // Label should now appear in the list
    await expect(page.getByText("Important")).toBeVisible({ timeout: 5000 });
  });
});
