import { test, expect, type Page } from "@playwright/test";
import pg from "pg";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount, createTestEmail, createTestDigest } from "./fixtures";

const DB = process.env.E2E_DATABASE_URL || "postgresql://taran:taran@localhost:5432/taran?sslmode=disable";

// Records whether "Getting started" is ever in the DOM, even for one frame.
// A unit test cannot see a hydration flash; this can.
async function watchForChecklist(page: Page) {
  await page.addInitScript(() => {
    (window as unknown as { __seen: boolean }).__seen = false;
    const check = () => {
      if (document.body && document.body.innerText.includes("Getting started")) {
        (window as unknown as { __seen: boolean }).__seen = true;
      }
    };
    new MutationObserver(check).observe(document, { childList: true, subtree: true });
    const t = setInterval(check, 10);
    setTimeout(() => clearInterval(t), 15000);
  });
}

async function everAppeared(page: Page) {
  return page.evaluate(() => (window as unknown as { __seen: boolean }).__seen);
}

async function setupCompleteAccount(userId: string) {
  const account = await createEmailAccount(userId);
  const item = await createTestEmail(userId, account, { subject: "A newsletter" });
  await createTestDigest(userId, [item]);
  // Configure preferences so the fourth item counts as done.
  const c = new pg.Client(DB);
  await c.connect();
  await c.query(
    `INSERT INTO user_preference (user_id, digest_email, digest_timezone, created_at, updated_at)
     VALUES ($1, true, 'Europe/London', NOW(), NOW())
     ON CONFLICT (user_id) DO UPDATE SET digest_email = true, digest_timezone = 'Europe/London'`,
    [userId]
  );
  await c.end();
}

test.describe("onboarding checklist on a finished account", () => {
  let userId = "";
  test.afterEach(async () => { if (userId) await cleanupTestUser(userId); });

  test("never appears, not even for a frame", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await setupCompleteAccount(user.id);

    await watchForChecklist(page);
    await page.goto("/");
    await expect(page.getByRole("heading", { name: "Dashboard", level: 1 })).toBeVisible({ timeout: 20000 });
    await page.waitForTimeout(4000); // let preferences resolve and any late render happen

    expect(await everAppeared(page)).toBe(false);
    await expect(page.getByText("Getting started")).toHaveCount(0);
    await page.screenshot({ path: "/tmp/ob-complete.png" });
  });

  test("still guides an account that has not finished", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(user.id); // inbox only: 1 of 4

    await page.goto("/");
    await expect(page.getByText("Getting started")).toBeVisible({ timeout: 20000 });
    await expect(page.getByText("1 of 4 complete")).toBeVisible();
    await page.screenshot({ path: "/tmp/ob-incomplete.png" });
  });
});
