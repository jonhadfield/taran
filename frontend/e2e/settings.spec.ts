import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import pg from "pg";

const DB_URL = process.env.DATABASE_URL || "postgresql://taran:taran@localhost:5432/taran?sslmode=disable";

let userId: string;

async function createEmailAccount(userId: string): Promise<void> {
  const client = new pg.Client(DB_URL);
  await client.connect();
  try {
    await client.query(
      `INSERT INTO email_account (id, user_id, email_address, display_name, is_active, created_at, updated_at)
       VALUES ($1, $2, $3, $4, true, NOW(), NOW())`,
      [crypto.randomUUID(), userId, `test-${userId.slice(0, 8)}@mailbrief.io`, "Test Inbox"]
    );
  } finally {
    await client.end();
  }
}

test.describe("Settings page", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("shows settings page with all sections", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings");

    await expect(page.getByRole("heading", { name: /Settings/i })).toBeVisible({ timeout: 10000 });

    // Check group headers
    await expect(page.getByText("General")).toBeVisible();
    await expect(page.getByText("Digest")).toBeVisible();
    await expect(page.getByText("Organisation")).toBeVisible();
    await expect(page.getByText("Advanced")).toBeVisible();
  });

  test("settings nav pill bar is visible", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings");

    // Nav pills should be visible
    await expect(page.getByRole("button", { name: "Accounts" })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole("button", { name: "Delivery" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Labels" })).toBeVisible();
  });

  test("theme color picker is visible", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings");

    await expect(page.getByText(/Accent Color|Appearance/i)).toBeVisible({ timeout: 10000 });
  });
});
