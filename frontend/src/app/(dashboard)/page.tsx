import { redirect } from "next/navigation";
import { serverFetch } from "@/lib/server-api";
import type { EmailAccount, ListResponse, DashboardData } from "@/types/api";
import { DashboardContent } from "./dashboard-content";

export default async function DashboardPage() {
  // Redirect to onboarding if user has no accounts
  // NOTE: redirect() must be called outside try/catch because Next.js
  // implements it by throwing a NEXT_REDIRECT error that must propagate.
  let hasAccount = false;
  let emailAddress = "";
  try {
    const accountRes = await serverFetch<ListResponse<EmailAccount>>("accounts");
    if (accountRes.data && accountRes.data.length > 0) {
      hasAccount = true;
      emailAddress = accountRes.data[0]?.EmailAddress ?? "";
    }
  } catch {
    // If fetch fails (e.g. not authenticated), let middleware handle it
  }

  if (!hasAccount) {
    redirect("/onboarding");
  }

  const emptyData: DashboardData = {
    Emails: [],
    Digests: [],
    UnreadCount: 0,
    Stats: { EmailsThisWeek: 0, EmailsLastWeek: 0, TotalEmails: 0, TopSenders: [] },
    Processing: { PendingCount: 0, ProcessingCount: 0, FailedCount: 0 },
    WeeklyHistory: [],
    TopicsWithCount: [],
    Categories: [],
    ActionItems: 0,
    Heatmap: [],
  };

  let initialData = emptyData;

  try {
    initialData = await serverFetch<DashboardData>("dashboard");
  } catch {
    // Will show empty state
  }

  return <DashboardContent initialData={initialData} emailAddress={emailAddress} />;
}
