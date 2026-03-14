import { isAdmin } from "@/lib/admin";
import { redirect } from "next/navigation";
import { AdminDashboard } from "./admin-dashboard";
import { AdminUsers } from "./admin-users";
import { AdminFailedEmails } from "./admin-failed-emails";
import { InviteForm } from "./invite-form";
import { WaitlistPanel } from "./waitlist-panel";

export default async function AdminPage() {
  const admin = await isAdmin();
  if (!admin) redirect("/");

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Admin Dashboard</h1>
      <AdminDashboard />
      <h2 className="text-xl font-bold pt-4">Pipeline & Failed Emails</h2>
      <AdminFailedEmails />
      <h2 className="text-xl font-bold pt-4">Users</h2>
      <AdminUsers />
      <h2 className="text-xl font-bold pt-4">Waitlist</h2>
      <WaitlistPanel />
      <h2 className="text-xl font-bold pt-4">Invitations</h2>
      <InviteForm />
    </div>
  );
}
