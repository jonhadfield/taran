import { isAdmin } from "@/lib/admin";
import { redirect } from "next/navigation";
import { AdminDashboard } from "./admin-dashboard";

export default async function AdminPage() {
  const admin = await isAdmin();
  if (!admin) redirect("/");

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Admin Dashboard</h1>
      <AdminDashboard />
    </div>
  );
}
