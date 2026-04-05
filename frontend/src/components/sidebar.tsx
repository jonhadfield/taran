"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { Home, Inbox, BookOpen, Settings, Users, Shield, PanelLeftClose, PanelLeftOpen } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
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

export function Sidebar({
  className,
  isAdmin,
  collapsed,
  onToggle,
}: {
  className?: string;
  isAdmin?: boolean;
  collapsed?: boolean;
  onToggle?: () => void;
}) {
  const pathname = usePathname();

  const isActive = (href: string) =>
    href === "/" ? pathname === "/" : pathname.startsWith(href);

  const navLink = (item: { href: string; label: string; icon: React.ComponentType<{ className?: string }> }) => {
    const link = (
      <Link
        key={item.href}
        href={item.href}
        className={cn(
          "flex items-center rounded-md text-sm font-medium transition-colors border-l-2",
          collapsed ? "justify-center px-2 py-2 gap-0" : "gap-3 px-3 py-2",
          isActive(item.href)
            ? "border-primary/60 bg-primary/10 text-foreground"
            : "border-transparent text-muted-foreground hover:bg-accent hover:text-foreground"
        )}
      >
        <item.icon className="size-4 shrink-0" />
        {!collapsed && item.label}
      </Link>
    );

    if (collapsed) {
      return (
        <Tooltip key={item.href}>
          <TooltipTrigger asChild>{link}</TooltipTrigger>
          <TooltipContent side="right">{item.label}</TooltipContent>
        </Tooltip>
      );
    }

    return link;
  };

  return (
    <div
      className={cn(
        "flex h-full flex-col border-r bg-card",
        className
      )}
    >
      {/* Branding */}
      <div className={cn("flex h-14 items-center gap-2.5", collapsed ? "justify-center px-2" : "px-5")}>
        <Image
          src="/logo.svg"
          alt={APP_NAME}
          width={32}
          height={32}
          className="rounded-lg shrink-0"
        />
        {!collapsed && <span className="font-semibold text-foreground">{APP_NAME}</span>}
      </div>

      {/* Main navigation */}
      <nav className={cn("flex flex-1 flex-col gap-1 pt-2", collapsed ? "px-1.5" : "px-3")}>
        {mainNav.map(navLink)}
      </nav>

      {/* Bottom navigation */}
      <nav className={cn("flex flex-col gap-1 border-t py-3", collapsed ? "px-1.5" : "px-3")}>
        {isAdmin && navLink({ href: "/admin", label: "Admin", icon: Shield })}
        {bottomNav.map(navLink)}
        {onToggle && (
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size={collapsed ? "icon" : "sm"}
                className={cn(
                  "text-muted-foreground hover:text-foreground",
                  collapsed ? "mx-auto" : "w-full justify-start gap-3 px-3"
                )}
                onClick={onToggle}
              >
                {collapsed ? (
                  <PanelLeftOpen className="size-4" />
                ) : (
                  <>
                    <PanelLeftClose className="size-4" />
                    <span className="text-sm font-medium">Collapse</span>
                  </>
                )}
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">
              {collapsed ? "Expand sidebar" : "Collapse sidebar"}
            </TooltipContent>
          </Tooltip>
        )}
      </nav>
    </div>
  );
}
