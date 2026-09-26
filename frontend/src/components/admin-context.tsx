"use client";

import { createContext, useContext } from "react";

/**
 * Admin status, resolved once on the server in the dashboard layout and shared
 * with the client components that need it.
 *
 * This only decides what is shown. The API enforces admin access on its own,
 * so a client that lies to itself gains nothing.
 */
const AdminContext = createContext(false);

export function AdminProvider({
  isAdmin,
  children,
}: {
  isAdmin: boolean;
  children: React.ReactNode;
}) {
  return <AdminContext.Provider value={isAdmin}>{children}</AdminContext.Provider>;
}

export function useIsAdmin(): boolean {
  return useContext(AdminContext);
}
