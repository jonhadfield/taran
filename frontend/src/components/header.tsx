"use client";

import { useSyncExternalStore } from "react";
import { authClient } from "@/lib/auth-client";
import { useRouter } from "next/navigation";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Sidebar } from "@/components/sidebar";
import { Menu, LogOut, Sun, Moon, Bell, BellOff, BellRing, Github, Bug } from "lucide-react";
import { useTheme } from "next-themes";
import { useEmailNotifications } from "@/hooks/use-email-notifications";
import { cn } from "@/lib/utils";

export function Header({ isAdmin }: { isAdmin?: boolean }) {
  const { data: session } = authClient.useSession();
  const router = useRouter();
  const { theme, setTheme } = useTheme();
  const { permission, requestPermission } = useEmailNotifications();
  const mounted = useSyncExternalStore(() => () => {}, () => true, () => false);

  const handleSignOut = async () => {
    await authClient.signOut();
    router.push("/login");
  };

  const user = session?.user;
  const initials = user?.name
    ? user.name
        .split(" ")
        .map((n) => n[0])
        .join("")
        .toUpperCase()
        .slice(0, 2)
    : "?";

  return (
    <header className="flex h-14 items-center justify-between border-b px-4 lg:px-6">
      <div className="flex items-center gap-3">
        {/* Mobile sidebar trigger */}
        <Sheet>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="lg:hidden">
              <Menu className="size-5" />
            </Button>
          </SheetTrigger>
          <SheetContent
            side="left"
            className="w-64 p-0"
          >
            <Sidebar isAdmin={isAdmin} />
          </SheetContent>
        </Sheet>
      </div>

      <div className="flex items-center gap-1">
        <a
          href="https://github.com/jonhadfield/taran"
          target="_blank"
          rel="noopener noreferrer"
          title="View on GitHub"
        >
          <Button variant="ghost" size="icon" asChild>
            <span>
              <Github className="size-4" />
              <span className="sr-only">View on GitHub</span>
            </span>
          </Button>
        </a>
        {mounted && permission !== "unsupported" && (
          <Button
            variant="ghost"
            size="icon"
            className={cn(
              "relative",
              permission === "default" && "text-foreground"
            )}
            onClick={requestPermission}
            title={
              permission === "granted"
                ? "Notifications enabled"
                : permission === "denied"
                  ? "Notifications blocked — update in browser settings"
                  : "Enable notifications"
            }
          >
            {permission === "granted" ? (
              <BellRing className="size-4" />
            ) : permission === "denied" ? (
              <BellOff className="size-4 text-muted-foreground" />
            ) : (
              <>
                <Bell className="size-4" />
                <span className="absolute top-1.5 right-1.5 size-2 rounded-full bg-blue-500 animate-pulse" />
              </>
            )}
            <span className="sr-only">Toggle notifications</span>
          </Button>
        )}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
        >
          <Sun className="size-4 rotate-0 scale-100 transition-transform dark:-rotate-90 dark:scale-0" />
          <Moon className="absolute size-4 rotate-90 scale-0 transition-transform dark:rotate-0 dark:scale-100" />
          <span className="sr-only">Toggle theme</span>
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="relative size-8 rounded-full">
              <Avatar className="size-8">
                <AvatarImage
                  src={user?.image || undefined}
                  alt={user?.name || "User"}
                />
                <AvatarFallback>{initials}</AvatarFallback>
              </Avatar>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <div className="px-2 py-1.5">
              <p className="text-sm font-medium">{user?.name || "User"}</p>
              <p className="text-xs text-muted-foreground">{user?.email}</p>
            </div>
            <DropdownMenuItem asChild>
              <a
                href="https://github.com/jonhadfield/taran/issues"
                target="_blank"
                rel="noopener noreferrer"
              >
                <Bug className="mr-2 size-4" />
                Report an issue
              </a>
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleSignOut}>
              <LogOut className="mr-2 size-4" />
              Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
