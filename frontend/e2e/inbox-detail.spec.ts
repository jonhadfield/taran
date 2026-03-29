import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount, createTestEmail } from "./fixtures";

let userId: string;

test.describe("Inbox detail page", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("shows email header, sender, and date", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account, {
      subject: "Header Test Email",
      fromName: "Alice Johnson",
      fromAddress: "alice@example.com",
    });

    await page.goto(`/inbox/${email.emailId}`);

    await expect(page.getByRole("heading", { name: /Header Test Email/i })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText("Alice Johnson")).toBeVisible();
    await expect(page.getByText("alice@example.com")).toBeVisible();
  });

  test("shows reading time estimate", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account, {
      subject: "Reading Time Test",
    });

    await page.goto(`/inbox/${email.emailId}`);

    await expect(page.getByText(/min read/i)).toBeVisible({ timeout: 10000 });
  });

  test("feedback buttons render for extracted emails", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account, {
      subject: "Feedback Test Email",
    });

    await page.goto(`/inbox/${email.emailId}`);

    await expect(page.getByText("Was this useful?")).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole("button", { name: /Yes/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /No/i })).toBeVisible();
  });

  test("delete button shows confirmation dialog", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account, {
      subject: "Delete Test Email",
    });

    await page.goto(`/inbox/${email.emailId}`);

    await expect(page.getByRole("heading", { name: /Delete Test Email/i })).toBeVisible({ timeout: 10000 });

    // Click delete
    await page.getByRole("button", { name: "Delete" }).click();

    // Confirmation dialog should appear
    await expect(page.getByText("Are you sure you want to delete")).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole("button", { name: "Cancel" })).toBeVisible();
  });

  test("breadcrumb navigates back to inbox", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account, {
      subject: "Breadcrumb Test",
    });

    await page.goto(`/inbox/${email.emailId}`);

    await expect(page.getByRole("heading", { name: /Breadcrumb Test/i })).toBeVisible({ timeout: 10000 });

    // Click the "Inbox" breadcrumb link (in main content, not sidebar)
    await page.getByRole("main").getByRole("link", { name: "Inbox" }).click();

    await expect(page).toHaveURL(/\/inbox$/);
  });
});
