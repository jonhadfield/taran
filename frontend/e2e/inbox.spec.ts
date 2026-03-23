import { test, expect } from "@playwright/test";
import { loginAsTestUser, cleanupTestUser } from "./auth-helper";
import { createEmailAccount, createTestEmail } from "./fixtures";

let userId: string;

test.describe("Inbox", () => {
  test.afterEach(async () => {
    if (userId) {
      await cleanupTestUser(userId);
    }
  });

  test("shows email in inbox list", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    await createTestEmail(userId, account, {
      subject: "Weekly AI Roundup",
      fromName: "Tech Newsletter",
    });

    await page.goto("/inbox");

    await expect(page.getByText("Weekly AI Roundup")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText("Tech Newsletter")).toBeVisible();
  });

  test("clicking email shows preview with AI summary", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    await createTestEmail(userId, account, {
      subject: "Deep Dive: Kubernetes",
      fromName: "DevOps Weekly",
      summary: "An in-depth look at Kubernetes best practices for production.",
      topics: ["Kubernetes", "DevOps"],
      keyPoints: ["Use resource limits", "Monitor pod health"],
    });

    await page.goto("/inbox");

    await expect(page.getByText("Deep Dive: Kubernetes")).toBeVisible({ timeout: 10000 });

    // Click the email — navigates to detail page
    await page.getByText("Deep Dive: Kubernetes").click();

    // Should show AI summary with extraction content
    await expect(page.getByText("AI Summary")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText("An in-depth look at Kubernetes best practices")).toBeVisible();
    await expect(page.getByText("Key Points")).toBeVisible();
  });

  test("star and archive actions work", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    const email = await createTestEmail(userId, account, {
      subject: "Action Test Email",
      fromName: "Test Sender",
    });

    // Navigate directly to the email detail page to avoid list re-render races
    await page.goto(`/inbox/${email.emailId}`);

    await expect(page.getByRole("heading", { name: /Action Test Email/i })).toBeVisible({ timeout: 10000 });

    // Star the email
    await page.getByRole("button", { name: "Star" }).click();
    await expect(page.getByRole("button", { name: "Starred" })).toBeVisible({ timeout: 5000 });

    // Archive the email
    await page.getByRole("button", { name: "Archive" }).click();
    await expect(page.getByRole("button", { name: "Unarchive" })).toBeVisible({ timeout: 5000 });
  });

  test("empty inbox shows illustration", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    await createEmailAccount(userId);

    await page.goto("/inbox");

    await expect(page.getByText("No emails yet")).toBeVisible({ timeout: 10000 });
  });

  test("unread email shows blue dot indicator", async ({ context, page }) => {
    const user = await loginAsTestUser(context);
    userId = user.id;
    const account = await createEmailAccount(userId);
    await createTestEmail(userId, account, {
      subject: "Unread Test Email",
      isRead: false,
    });

    await page.goto("/inbox");

    await expect(page.getByText("Unread Test Email")).toBeVisible({ timeout: 10000 });

    // Sender name should be bold (font-semibold) for unread
    const senderEl = page.getByText("Test Sender").first();
    await expect(senderEl).toHaveClass(/font-semibold/);
  });
});
