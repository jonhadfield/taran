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

  try {
    const res = await serverFetch<ListResponse<Email>>(
      `emails?${queryString}`
    );
    emails = res.data || [];
    total = res.total;
  } catch {
    // Will show empty state
  }

  return (
    <InboxList
      initialEmails={emails}
      initialTotal={total}
      filter={filter}
      queryString={queryString}
    />
  );
}
