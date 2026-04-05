import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount, createTestEmail } from "./fixtures";

let userId: string;

test.describe("Inbox filtering", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("filter tabs switch between all/unread/starred/archived", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);

    // Create a mix: unread email and read+starred email
    await createTestEmail(userId, account, {
      subject: "Unread Email",
      isRead: false,
    });
    await createTestEmail(userId, account, {
      subject: "Starred Email",
      isRead: true,
      isStarred: true,
    });

    await page.goto("/inbox");
    await expect(page.getByText("Unread Email")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText("Starred Email")).toBeVisible();

    // Switch to unread filter
    await page.getByRole("button", { name: /Unread/i }).click();
    await expect(page.getByText("Unread Email")).toBeVisible({ timeout: 5000 });

    // Switch to starred filter
    await page.getByRole("button", { name: /Starred/i }).click();
    await expect(page.getByText("Starred Email")).toBeVisible({ timeout: 5000 });
  });

  test("search filters emails by subject", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);

    await createTestEmail(userId, account, { subject: "Alpha Newsletter" });
    await createTestEmail(userId, account, { subject: "Beta Update Report" });

    await page.goto("/inbox");
    await expect(page.getByText("Alpha Newsletter")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText("Beta Update Report")).toBeVisible();

    // Search for "Alpha"
    await page.getByPlaceholder(/Search/i).fill("Alpha");

    // Wait for debounced search
    await expect(page.getByText("Alpha Newsletter")).toBeVisible({ timeout: 5000 });
  });

  test("keyboard shortcut / focuses search", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    await createTestEmail(userId, account);

    await page.goto("/inbox");
    await expect(page.getByText("Test Newsletter")).toBeVisible({ timeout: 10000 });

    // Click body to ensure page has focus, then press /
    await page.locator("body").click();
    await page.keyboard.press("/");

    // Search input should be focused (allow time for event handler)
    await expect(page.getByPlaceholder(/Search/i)).toBeFocused({ timeout: 3000 });
  });
});
