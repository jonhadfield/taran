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

  test("create a new label via settings", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings");
    await expect(page.getByRole("heading", { name: /Settings/i })).toBeVisible({ timeout: 10000 });

    // Click "New Label" button
    const newLabelButton = page.getByRole("button", { name: /New Label/i });
    await newLabelButton.scrollIntoViewIfNeeded();
    await newLabelButton.click();

    // Fill in label name
    await expect(page.getByPlaceholder("Label name")).toBeVisible({ timeout: 5000 });
    await page.getByPlaceholder("Label name").fill("Important");

    // Submit with Enter
    await page.getByPlaceholder("Label name").press("Enter");

    // Label should appear in the list
    await expect(page.getByText("Important")).toBeVisible({ timeout: 5000 });
  });
});
