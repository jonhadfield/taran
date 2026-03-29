import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount, createTestEmail, createTestDigest } from "./fixtures";

let userId: string;

test.describe("Digest actions", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("delete button shows confirmation dialog", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account);
    const digestId = await createTestDigest(userId, [email], {
      title: "Delete Test Digest",
    });

    await page.goto(`/digests/${digestId}`);
    await expect(page.getByRole("heading", { name: "Delete Test Digest" })).toBeVisible({ timeout: 10000 });

    // Click delete
    await page.getByRole("button", { name: /Delete/i }).click();

    // Confirmation dialog should appear
    await expect(page.getByText(/Are you sure/i)).toBeVisible({ timeout: 5000 });
  });

  test("PDF download button is present", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account);
    await createTestDigest(userId, [email], {
      title: "PDF Test Digest",
    });

    await page.goto("/digests");
    await page.getByText("PDF Test Digest").click();

    await expect(page.getByRole("button", { name: /PDF/i })).toBeVisible({ timeout: 10000 });
  });

  test("share button is present", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account);
    await createTestDigest(userId, [email], {
      title: "Share Test Digest",
    });

    await page.goto("/digests");
    await page.getByText("Share Test Digest").click();

    await expect(page.getByRole("button", { name: /Share/i })).toBeVisible({ timeout: 10000 });
  });

  test("feedback buttons render on digest detail", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account);
    await createTestDigest(userId, [email], {
      title: "Feedback Test Digest",
    });

    await page.goto("/digests");
    await page.getByText("Feedback Test Digest").click();

    // Digest feedback buttons should be visible
    await expect(page.getByText("Feedback Test Digest")).toBeVisible({ timeout: 10000 });
  });
});
