import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount, createTestEmail, createTestDigest } from "./fixtures";
import pg from "pg";

const DB_URL = process.env.E2E_DATABASE_URL || "postgresql://taran:taran@localhost:5432/taran?sslmode=disable";

let userId: string;

test.describe("Shared digest (public)", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("shows shared digest without authentication", async ({ page }) => {
    // Create test data (no browser context needed for DB setup)
    const context = await page.context();
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account);
    const digestId = await createTestDigest(userId, [email]);

    // Generate a share token directly in DB
    const shareToken = "test-share-token-" + Date.now();
    const client = new pg.Client(DB_URL);
    await client.connect();
    try {
      await client.query(`UPDATE digest SET share_token = $1 WHERE id = $2`, [shareToken, digestId]);
    } finally {
      await client.end();
    }

    // Open shared digest in a fresh context (no auth cookies)
    const freshPage = await (await page.context().browser()!.newContext()).newPage();
    await freshPage.goto(`/shared/${shareToken}`);

    // Should show digest content without requiring login
    await expect(freshPage.getByText("Weekly Digest")).toBeVisible({ timeout: 10000 });
    await freshPage.close();
  });

  test("shows 404 for invalid share token", async ({ page }) => {
    await page.goto("/shared/nonexistent-token-12345");

    // Should show not found
    await expect(page.getByText(/not found|404/i)).toBeVisible({ timeout: 10000 });
  });
});
