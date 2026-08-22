import { redirect } from "next/navigation";
import { serverFetch } from "@/lib/server-api";
import type { EmailAccount, ListResponse, DashboardData } from "@/types/api";
import { DashboardContent } from "./dashboard-content";

export default async function DashboardPage() {
  const emptyData: DashboardData = {
    Emails: [],
    Digests: [],
    UnreadCount: 0,
    Stats: { EmailsThisWeek: 0, EmailsLastWeek: 0, TotalEmails: 0, TopSenders: [] },
    Processing: { PendingCount: 0, ProcessingCount: 0, FailedCount: 0 },
    WeeklyHistory: [],
    TopicsWithCount: [],
    Categories: [],
    Heatmap: [],
  };

  // Start the dashboard aggregate immediately rather than after the account
  // lookup resolves. The two are independent, and awaiting them in sequence
  // added a full backend round-trip to every dashboard load. The catch is
  // attached here so the promise can never reject unhandled if the account
  // check redirects away before it is awaited.
  const dashboardRequest = serverFetch<DashboardData>("dashboard").catch(
    () => emptyData,
  );

  // Redirect to onboarding if user has no accounts.
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

  const initialData = await dashboardRequest;

  return <DashboardContent initialData={initialData} emailAddress={emailAddress} />;
}
