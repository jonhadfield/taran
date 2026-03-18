import { type BrowserContext } from "@playwright/test";
import { randomUUID, createHmac } from "crypto";
import pg from "pg";

const DB_URL = process.env.DATABASE_URL || "postgresql://taran:taran@localhost:5432/taran?sslmode=disable";
const AUTH_SECRET = process.env.BETTER_AUTH_SECRET || "test-secret";

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
  } finally {
    await client.end();
  }

  // Better Auth cookie format: token.hmac_signature
  const signature = createHmac("sha256", AUTH_SECRET).update(token).digest("hex");
  const cookieValue = `${token}.${signature}`;

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
    // Cascade deletes will handle sessions, accounts, emails, etc.
    await client.query(`DELETE FROM "user" WHERE id = $1`, [userId]);
  } finally {
    await client.end();
  }
}
