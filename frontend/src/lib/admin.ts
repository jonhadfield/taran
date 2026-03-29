import { headers } from "next/headers";
import { auth } from "@/lib/auth";

export async function isAdmin(): Promise<boolean> {
  const adminEmails = process.env.ADMIN_EMAILS;
  if (!adminEmails) return false;

  const session = await auth.api.getSession({
    headers: await headers(),
  });

  if (!session?.user?.email) return false;

  const adminList = adminEmails.split(",").map((e) => e.trim().toLowerCase());
  return adminList.includes(session.user.email.toLowerCase());
}
