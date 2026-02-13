import { Header } from "@/components/header";
import { Sidebar } from "@/components/sidebar";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen">
      {/* Fixed full-height sidebar - hidden on mobile */}
      <aside className="hidden lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col">
        <Sidebar />
      </aside>

      {/* Main content area offset by sidebar width */}
      <div className="lg:pl-64">
        <Header />
        <main className="@container p-4 lg:p-6">{children}</main>
      </div>
    </div>
  );
}
