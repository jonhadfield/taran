"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { Home, Inbox, BookOpen, Settings, Users, Shield } from "lucide-react";
import Image from "next/image";
import { APP_NAME } from "@/lib/config";

const mainNav = [
  { href: "/", label: "Dashboard", icon: Home },
  { href: "/inbox", label: "Inbox", icon: Inbox },
  { href: "/digests", label: "Digests", icon: BookOpen },
  { href: "/senders", label: "Senders", icon: Users },
];

const bottomNav = [
  { href: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar({ className, isAdmin }: { className?: string; isAdmin?: boolean }) {
  const pathname = usePathname();

  const isActive = (href: string) =>
    href === "/" ? pathname === "/" : pathname.startsWith(href);

  const navLink = (item: { href: string; label: string; icon: React.ComponentType<{ className?: string }> }) => (
    <Link
      key={item.href}
      href={item.href}
      className={cn(
        "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors border-l-2",
        isActive(item.href)
          ? "border-primary/60 bg-primary/10 text-foreground"
          : "border-transparent text-muted-foreground hover:bg-accent hover:text-foreground"
      )}
    >
      <item.icon className="size-4" />
      {item.label}
    </Link>
  );

  return (
    <div
      className={cn(
        "flex h-full flex-col border-r bg-card",
        className
      )}
    >
      {/* Branding */}
      <div className="flex h-14 items-center gap-2.5 px-5">
        <Image
          src="/logo.svg"
          alt={APP_NAME}
          width={32}
          height={32}
          className="rounded-lg"
        />
        <span className="font-semibold text-foreground">{APP_NAME}</span>
      </div>

      {/* Main navigation */}
      <nav className="flex flex-1 flex-col gap-1 px-3 pt-2">
        {mainNav.map(navLink)}
      </nav>

      {/* Bottom navigation */}
      <nav className="flex flex-col gap-1 border-t px-3 py-3">
        {isAdmin && navLink({ href: "/admin", label: "Admin", icon: Shield })}
        {bottomNav.map(navLink)}
      </nav>
    </div>
  );
}
