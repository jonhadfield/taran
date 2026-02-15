import { redirect } from "next/navigation";
import { serverFetch } from "@/lib/server-api";
import type { Email, Digest, EmailAccount, ListResponse } from "@/types/api";
import { DashboardContent } from "./dashboard-content";

export default async function DashboardPage() {
  // Redirect to setup if user has no accounts
  try {
    const accountRes = await serverFetch<ListResponse<EmailAccount>>("accounts");
    if (!accountRes.data || accountRes.data.length === 0) {
      redirect("/setup");
    }
  } catch {
    // If fetch fails (e.g. not authenticated), let middleware handle it
  }

  let emails: Email[] = [];
  let emailTotal = 0;
  let digests: Digest[] = [];
  let unreadCount = 0;

  try {
    const [emailRes, digestRes, unreadRes] = await Promise.all([
      serverFetch<ListResponse<Email>>("emails?limit=5"),
      serverFetch<ListResponse<Digest>>("digests?limit=3"),
      serverFetch<ListResponse<Email>>("emails?is_read=false&limit=1"),
    ]);
    emails = emailRes.data || [];
    emailTotal = emailRes.total;
    digests = digestRes.data || [];
    unreadCount = unreadRes.total;
  } catch {
    // Will show empty state
  }

  return (
    <DashboardContent
      initialEmails={emails}
      initialEmailTotal={emailTotal}
      initialDigests={digests}
      initialUnreadCount={unreadCount}
    />
  );
}
