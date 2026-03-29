import { serverFetch } from "@/lib/server-api";
import type { Email, EmailAccount, ListResponse, UserPreference } from "@/types/api";
import { InboxList } from "./inbox-list";

export default async function InboxPage({
  searchParams,
}: {
  searchParams: Promise<{ filter?: string }>;
}) {
  const { filter = "all" } = await searchParams;

  const params: string[] = ["limit=50"];
  if (filter === "unread") params.push("is_read=false");
  if (filter === "starred") params.push("is_starred=true");
  if (filter === "archived") params.push("is_archived=true");

  const queryString = params.join("&");

  let emails: Email[] = [];
  let total = 0;
  let topics: string[] = [];
  let emailAddress = "";

  try {
    const prefRes = await serverFetch<UserPreference>("preferences").catch(() => null);
    const topicLimit = prefRes?.TopicLimit ?? 15;

    const [emailRes, topicsRes, accountRes] = await Promise.all([
      serverFetch<ListResponse<Email>>(`emails?${queryString}`),
      serverFetch<string[]>(`topics?limit=${topicLimit}`).catch(() => [] as string[]),
      serverFetch<ListResponse<EmailAccount>>("accounts").catch(() => ({ data: [], total: 0 }) as ListResponse<EmailAccount>),
    ]);
    emails = emailRes.data || [];
    total = emailRes.total;
    topics = topicsRes;
    emailAddress = accountRes.data?.[0]?.EmailAddress ?? "";
  } catch {
    // Will show empty state
  }

  return (
    <InboxList
      initialEmails={emails}
      initialTotal={total}
      filter={filter}
      queryString={queryString}
      initialTopics={topics}
      emailAddress={emailAddress}
    />
  );
}
