import { type BrowserContext } from "@playwright/test";
import { randomUUID, createHmac } from "crypto";
import pg from "pg";

// E2E tests always use the local test database. Use E2E_DATABASE_URL to override,
// NOT DATABASE_URL (which typically points to production via .env).
const DB_URL = process.env.E2E_DATABASE_URL || "postgresql://taran:taran@localhost:5432/taran?sslmode=disable";
// No fallback: forged session cookies must be signed with the same secret the
// app verifies with, so a missing value has to fail loudly rather than silently
// sign with a hardcoded one. Playwright loads .env/.env.local before this runs.
const AUTH_SECRET = ((): string => {
  const secret = process.env.BETTER_AUTH_SECRET;
  if (!secret) {
    throw new Error(
      "BETTER_AUTH_SECRET is not set. E2E tests forge session cookies and must " +
        "sign them with the secret the app verifies with. Set it in " +
        "frontend/.env.local, or in the environment for CI."
    );
  }
  return secret;
})();

interface TestUser {
  id: string;
  email: string;
  name: string;
  token: string;
}

/**
 * Create a test user and session directly in the database, then set the
 * session cookie on the browser context. This bypasses OAuth entirely.
 */
export async function loginAsTestUser(
  context: BrowserContext,
  opts?: { email?: string; name?: string }
): Promise<TestUser> {
  const userId = randomUUID();
  const sessionId = randomUUID();
  const token = randomUUID();
  const email = opts?.email || `test-${userId.slice(0, 8)}@example.com`;
  const name = opts?.name || "Test User";

  const client = new pg.Client(DB_URL);
  await client.connect();

  try {
    // Create user
    await client.query(
      `INSERT INTO "user" (id, email, name, "emailVerified", "createdAt", "updatedAt")
       VALUES ($1, $2, $3, true, NOW(), NOW())`,
      [userId, email, name]
    );

    // Create session (30 days from now)
    await client.query(
      `INSERT INTO session (id, "userId", token, "expiresAt", "createdAt", "updatedAt")
       VALUES ($1, $2, $3, NOW() + INTERVAL '30 days', NOW(), NOW())`,
      [sessionId, userId, token]
    );

    // Create invite so the access check passes
    await client.query(
      `INSERT INTO invite (id, email, invited_by, created_at) VALUES ($1, $2, $3, NOW())
       ON CONFLICT (email) DO NOTHING`,
      [randomUUID(), email, userId]
    );
  } finally {
    await client.end();
  }

  // Better Auth (via better-call) cookie format: encodeURIComponent("token.base64Signature")
  // Signature is HMAC-SHA256(secret_utf8_bytes, token) → standard base64.
  const signature = createHmac("sha256", AUTH_SECRET).update(token).digest("base64");
  const cookieValue = encodeURIComponent(`${token}.${signature}`);

  await context.addCookies([
    {
      name: "better-auth.session_token",
      value: cookieValue,
      domain: "localhost",
      path: "/",
      httpOnly: true,
      sameSite: "Lax",
    },
  ]);

  return { id: userId, email, name, token };
}

/**
 * Clean up test user and all associated data.
 */
export async function cleanupTestUser(userId: string): Promise<void> {
  const client = new pg.Client(DB_URL);
  await client.connect();

  try {
    // Look up email to clean up invite
    const userRow = await client.query(`SELECT email FROM "user" WHERE id = $1`, [userId]);
    if (userRow.rows.length > 0) {
      await client.query(`DELETE FROM invite WHERE email = $1`, [userRow.rows[0].email]);
    }
    // Delete in dependency order — Better Auth tables don't have ON DELETE CASCADE
    await client.query(`DELETE FROM session WHERE "userId" = $1`, [userId]);
    await client.query(`DELETE FROM account WHERE "userId" = $1`, [userId]);
    // email_account CASCADE will delete emails → extractions, feedback, attachments, labels, digest_items
    await client.query(`DELETE FROM email_account WHERE user_id = $1`, [userId]);
    // Clean up digests (not cascade-linked to email_account)
    await client.query(`DELETE FROM digest WHERE user_id = $1`, [userId]);
    // Clean up user preferences, sender prefs, etc.
    await client.query(`DELETE FROM user_preference WHERE user_id = $1`, [userId]);
    await client.query(`DELETE FROM sender_preference WHERE user_id = $1`, [userId]);
    await client.query(`DELETE FROM token_usage WHERE user_id = $1`, [userId]);
    await client.query(`DELETE FROM saved_search WHERE user_id = $1`, [userId]);
    await client.query(`DELETE FROM auto_archive_rule WHERE user_id = $1`, [userId]);
    await client.query(`DELETE FROM label WHERE user_id = $1`, [userId]);
    await client.query(`DELETE FROM user_llm_key WHERE user_id = $1`, [userId]);
    // Finally delete the user
    await client.query(`DELETE FROM "user" WHERE id = $1`, [userId]);
  } finally {
    await client.end();
  }
}
