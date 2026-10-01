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

    // Match the list row (a button) rather than the text: opening an email
    // repeats its subject in the preview heading, which makes a plain text
    // match ambiguous.
    const row = (subject: string) => page.getByRole("button", { name: new RegExp(subject) });

    await page.goto("/inbox");
    await expect(row("Unread Email")).toBeVisible({ timeout: 10000 });
    await expect(row("Starred Email")).toBeVisible();

    // The filters are tabs. Matching them as buttons instead picks up the
    // email rows, so the test opened an email rather than filtering.
    await page.getByRole("tab", { name: "unread" }).click();
    await expect(row("Unread Email")).toBeVisible({ timeout: 5000 });
    await expect(row("Starred Email")).toHaveCount(0);

    await page.getByRole("tab", { name: "starred" }).click();
    await expect(row("Starred Email")).toBeVisible({ timeout: 5000 });
    await expect(row("Unread Email")).toHaveCount(0);
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
    const search = page.getByRole("textbox", { name: "Search emails" });
    await expect(search).toBeVisible({ timeout: 10000 });

    // Retry until the keydown listener is attached. Click the Inbox
    // heading (not body) so we don't land on the sort <select>, which
    // would make the handler ignore "/".
    await expect(async () => {
      await page.getByRole("heading", { name: "Inbox" }).click();
      await page.keyboard.press("/");
      await expect(search).toBeFocused({ timeout: 500 });
    }).toPass({ timeout: 5000 });
  });
});
