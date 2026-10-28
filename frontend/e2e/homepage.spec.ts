import { test, expect } from "@playwright/test";

test.describe("Public homepage", () => {
  test("serves marketing content at / without a session", async ({ page }) => {
    await page.goto("/");

    await expect(page).toHaveURL("http://localhost:3002/");
    await expect(page).toHaveTitle("MailBrief — AI newsletter digests");
    await expect(
      page.getByRole("heading", { name: /Forward your newsletters/i }),
    ).toBeVisible();
    await expect(page.getByRole("link", { name: /Get started/i })).toHaveAttribute(
      "href",
      "/login",
    );
    await page.getByRole("link", { name: /Get started/i }).click();
    await expect(page).toHaveURL(/\/login/);
    await expect(page.getByRole("button", { name: /Continue with Google/i })).toBeVisible();
  });

  test("redirects /home to the canonical homepage", async ({ page }) => {
    await page.goto("/home");

    await expect(page).toHaveURL("http://localhost:3002/");
    await expect(
      page.getByRole("heading", { name: /Forward your newsletters/i }),
    ).toBeVisible();
  });
});
