import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount } from "./fixtures";

let userId: string;

test.describe("Command Palette", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("opens with Ctrl+K and shows commands", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/");
    await expect(page.getByRole("heading", { name: /Dashboard/i })).toBeVisible({ timeout: 10000 });

    // Open command palette
    await page.keyboard.press("Control+k");

    // Should show the palette with search input
    await expect(page.getByPlaceholder(/command/i)).toBeVisible({ timeout: 5000 });

    // Should show navigation group inside the dialog
    const dialog = page.getByRole("dialog");
    await expect(dialog.getByText("Navigation")).toBeVisible();
    await expect(dialog.getByText("Inbox")).toBeVisible();
    await expect(dialog.getByText("Digests")).toBeVisible();
  });

  test("filters commands by search", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/");
    await expect(page.getByRole("heading", { name: /Dashboard/i })).toBeVisible({ timeout: 10000 });

    await page.keyboard.press("Control+k");
    await expect(page.getByPlaceholder(/command/i)).toBeVisible({ timeout: 5000 });

    // Type to filter
    await page.getByPlaceholder(/command/i).fill("settings");

    // Settings should be visible, other items hidden
    await expect(page.getByRole("option", { name: /Settings/i })).toBeVisible();
  });

  test("navigates on selection", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/");
    await expect(page.getByRole("heading", { name: /Dashboard/i })).toBeVisible({ timeout: 10000 });

    await page.keyboard.press("Control+k");
    await expect(page.getByPlaceholder(/command/i)).toBeVisible({ timeout: 5000 });

    // Type and select inbox
    await page.getByPlaceholder(/command/i).fill("inbox");
    await page.keyboard.press("Enter");

    await expect(page).toHaveURL(/\/inbox/, { timeout: 5000 });
  });

  test("closes with Escape", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/");
    await expect(page.getByRole("heading", { name: /Dashboard/i })).toBeVisible({ timeout: 10000 });

    await page.keyboard.press("Control+k");
    await expect(page.getByPlaceholder(/command/i)).toBeVisible({ timeout: 5000 });

    await page.keyboard.press("Escape");

    await expect(page.getByPlaceholder(/command/i)).not.toBeVisible();
  });
});
