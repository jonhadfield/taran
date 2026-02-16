import { redirect } from "next/navigation";
import { serverFetch } from "@/lib/server-api";
import type { EmailAccount, ListResponse, DashboardData } from "@/types/api";
import { DashboardContent } from "./dashboard-content";

export default async function DashboardPage() {
  // Redirect to onboarding if user has no accounts
  try {
    const accountRes = await serverFetch<ListResponse<EmailAccount>>("accounts");
    if (!accountRes.data || accountRes.data.length === 0) {
      redirect("/onboarding");
    }
  } catch {
    // If fetch fails (e.g. not authenticated), let middleware handle it
  }

  const emptyData: DashboardData = {
    emails: [],
    emailTotal: 0,
    digests: [],
    unreadCount: 0,
    stats: { EmailsThisWeek: 0, EmailsLastWeek: 0, TotalEmails: 0, TopSenders: [] },
  };

  let initialData = emptyData;

  try {
    initialData = await serverFetch<DashboardData>("dashboard");
  } catch {
    // Will show empty state
  }

  return <DashboardContent initialData={initialData} />;
}
