import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount } from "./fixtures";

let userId: string;

test.describe("Settings page", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("the hub lists every group", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings");

    await expect(page.getByRole("heading", { name: /Settings/i })).toBeVisible({ timeout: 10000 });
    const hub = page.getByRole("navigation", { name: "Settings groups" });
    for (const group of ["Inbox", "Digest", "Organisation", "Account"]) {
      await expect(hub.getByRole("link", { name: new RegExp(`^${group}`) })).toBeVisible();
    }
  });

  test("a group opens from the hub and the rail stays visible", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings");
    await page.getByRole("navigation", { name: "Settings groups" }).getByRole("link", { name: /^Digest/ }).click();

    await expect(page).toHaveURL(/\/settings\/digest$/);
    await expect(page.getByRole("heading", { name: "Digest", level: 1 })).toBeVisible({ timeout: 10000 });

    // Every group is reachable without scrolling sideways.
    const rail = page.getByRole("navigation", { name: "Settings" });
    for (const group of ["Inbox", "Digest", "Organisation", "Account"]) {
      await expect(rail.getByRole("link", { name: group, exact: true })).toBeVisible();
    }
  });

  test("theme color picker is visible", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/settings/account");

    await expect(page.getByText("Accent color")).toBeVisible({ timeout: 10000 });
  });
});
