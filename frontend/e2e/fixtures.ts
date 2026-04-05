import pg from "pg";

const DB_URL = process.env.E2E_DATABASE_URL || "postgresql://taran:taran@localhost:5432/taran?sslmode=disable";

async function withClient<T>(fn: (client: pg.Client) => Promise<T>): Promise<T> {
  const client = new pg.Client(DB_URL);
  await client.connect();
  try {
    return await fn(client);
  } finally {
    await client.end();
  }
}

export async function createEmailAccount(
  userId: string,
  opts?: { username?: string }
): Promise<{ id: string; emailAddress: string }> {
  const id = crypto.randomUUID();
  const username = opts?.username || `test-${id.slice(0, 8)}`;
  const emailAddress = `${username}@mailbrief.io`;

  await withClient((client) =>
    client.query(
      `INSERT INTO email_account (id, user_id, email_address, display_name, is_active, created_at, updated_at)
       VALUES ($1, $2, $3, $4, true, NOW(), NOW())`,
      [id, userId, emailAddress, username]
    )
  );

  return { id, emailAddress };
}

export async function createTestEmail(
  userId: string,
  account: { id: string; emailAddress: string },
  opts?: {
    subject?: string;
    fromName?: string;
    fromAddress?: string;
    summary?: string;
    topics?: string[];
    keyPoints?: string[];
    isRead?: boolean;
    isStarred?: boolean;
  }
): Promise<{ emailId: string; extractionId: string }> {
  const emailId = crypto.randomUUID();
  const extractionId = crypto.randomUUID();
  const accountId = account.id;
  const emailAddress = account.emailAddress;

  const subject = opts?.subject || "Test Newsletter: AI and Tech Updates";
  const fromName = opts?.fromName || "Test Sender";
  const fromAddress = opts?.fromAddress || "sender@example.com";
  const summary = opts?.summary || "A comprehensive overview of recent developments in AI and technology.";
  const topics = opts?.topics || ["AI", "Technology", "News"];
  const keyPoints = opts?.keyPoints || ["AI adoption is accelerating", "New open-source models released"];

  await withClient(async (client) => {
    await client.query(
      `INSERT INTO email (id, user_id, email_account_id, message_id, from_address, from_name, to_address, subject, text_body, received_at, date_header, status, is_read, is_starred, is_archived)
       VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW(), 'processed', $10, $11, false)`,
      [
        emailId, userId, accountId, `<${emailId}@test>`,
        fromAddress, fromName, emailAddress, subject,
        `Text body for: ${subject}`,
        opts?.isRead ?? false, opts?.isStarred ?? false,
      ]
    );

    await client.query(
      `INSERT INTO extraction (id, email_id, summary, key_points, topics, links, action_items, sentiment, source_category, provider, model, tokens_used, processed_at)
       VALUES ($1, $2, $3, $4, $5, '[]'::jsonb, '[]'::jsonb, 'informational', 'newsletter', 'test', 'test', 0, NOW())`,
      [extractionId, emailId, summary, JSON.stringify(keyPoints), JSON.stringify(topics)]
    );
  });

  return { emailId, extractionId };
}

export async function createTestDigest(
  userId: string,
  items: { emailId: string; extractionId: string }[],
  opts?: { title?: string; summary?: string }
): Promise<string> {
  const digestId = crypto.randomUUID();
  const title = opts?.title || "Weekly Digest — Test";
  const summary = opts?.summary || "A test digest summarising recent emails.";

  await withClient(async (client) => {
    await client.query(
      `INSERT INTO digest (id, user_id, title, summary, highlights, top_topics, period_start, period_end, period_type, email_count, provider, model)
       VALUES ($1, $2, $3, $4, $5, $6, NOW() - INTERVAL '7 days', NOW(), 'weekly', $7, 'test', 'test')`,
      [
        digestId, userId, title, summary,
        JSON.stringify(["AI developments continue", "New tools released"]),
        JSON.stringify(["AI", "Technology"]),
        items.length,
      ]
    );

    for (let i = 0; i < items.length; i++) {
      await client.query(
        `INSERT INTO digest_item (id, digest_id, email_id, extraction_id, sort_order)
         VALUES ($1, $2, $3, $4, $5)`,
        [crypto.randomUUID(), digestId, items[i].emailId, items[i].extractionId, i]
      );
    }
  });

  return digestId;
}
