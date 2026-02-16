import { serverFetch } from "@/lib/server-api";
import type { Email, ListResponse } from "@/types/api";
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

  try {
    const [emailRes, topicsRes] = await Promise.all([
      serverFetch<ListResponse<Email>>(`emails?${queryString}`),
      serverFetch<string[]>("topics").catch(() => [] as string[]),
    ]);
    emails = emailRes.data || [];
    total = emailRes.total;
    topics = topicsRes;
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
    />
  );
}
