import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import pg from "pg";

const DB_URL = process.env.DATABASE_URL || "postgresql://taran:taran@localhost:5432/taran?sslmode=disable";

let userId: string;

async function createEmailAccount(userId: string): Promise<string> {
  const client = new pg.Client(DB_URL);
  await client.connect();
  try {
    const accountId = crypto.randomUUID();
    await client.query(
      `INSERT INTO email_account (id, user_id, email_address, display_name, is_active, created_at, updated_at)
       VALUES ($1, $2, $3, $4, true, NOW(), NOW())`,
      [accountId, userId, `test-${accountId.slice(0, 8)}@mailbrief.io`, "Test Inbox"]
    );
    return accountId;
  } finally {
    await client.end();
  }
}

test.describe("Dashboard", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("shows dashboard for users with an account", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/");

    await expect(page.getByRole("heading", { name: /Dashboard/i })).toBeVisible({ timeout: 10000 });
  });

  test("shows empty state when no emails", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/");

    await expect(page.getByText(/No emails yet/i)).toBeVisible({ timeout: 10000 });
  });

  test("sidebar navigation works", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/");
    await expect(page.getByRole("heading", { name: /Dashboard/i })).toBeVisible({ timeout: 10000 });

    // Navigate to inbox via sidebar
    await page.getByRole("link", { name: /Inbox/i }).first().click();
    await expect(page).toHaveURL(/\/inbox/);

    // Navigate to digests
    await page.getByRole("link", { name: /Digests/i }).first().click();
    await expect(page).toHaveURL(/\/digests/);

    // Navigate to senders
    await page.getByRole("link", { name: /Senders/i }).first().click();
    await expect(page).toHaveURL(/\/senders/);

    // Navigate to settings
    await page.getByRole("link", { name: /Settings/i }).first().click();
    await expect(page).toHaveURL(/\/settings/);
  });
});
