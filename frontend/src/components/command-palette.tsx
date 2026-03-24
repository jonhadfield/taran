"use client";

import { useCallback, useEffect, useSyncExternalStore, useState } from "react";
import { useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { toast } from "sonner";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@/components/ui/command";
import {
  Home,
  Inbox,
  BookOpen,
  Users,
  Settings,
  Shield,
  Moon,
  Sun,
  Download,
  Sparkles,
  Mail,
  Star,
  Archive,
  LinkIcon,
} from "lucide-react";

interface CommandPaletteProps {
  isAdmin: boolean;
}

export function CommandPalette({ isAdmin }: CommandPaletteProps) {
  const [open, setOpen] = useState(false);
  const router = useRouter();
  const { theme, setTheme } = useTheme();
  const mounted = useSyncExternalStore(() => () => {}, () => true, () => false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setOpen((prev) => !prev);
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  const runCommand = useCallback(
    (command: () => void) => {
      setOpen(false);
      command();
    },
    [],
  );

  const handleExport = useCallback(async () => {
    try {
      const res = await fetch("/api/proxy/export", { credentials: "include" });
      if (!res.ok) throw new Error("Export failed");
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "mailbrief-export.json";
      a.click();
      URL.revokeObjectURL(url);
      toast.success("Data exported");
    } catch {
      toast.error("Failed to export data");
    }
  }, []);

  return (
    <CommandDialog open={open} onOpenChange={setOpen} showCloseButton={false}>
      <CommandInput placeholder="Type a command or search..." />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>

        <CommandGroup heading="Navigation">
          <CommandItem onSelect={() => runCommand(() => router.push("/"))}>
            <Home />
            <span>Dashboard</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/inbox"))}>
            <Inbox />
            <span>Inbox</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/digests"))}>
            <BookOpen />
            <span>Digests</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/senders"))}>
            <Users />
            <span>Senders</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/senders/subscriptions"))}>
            <LinkIcon />
            <span>Subscriptions</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/settings"))}>
            <Settings />
            <span>Settings</span>
          </CommandItem>
          {isAdmin && (
            <CommandItem onSelect={() => runCommand(() => router.push("/admin"))}>
              <Shield />
              <span>Admin</span>
            </CommandItem>
          )}
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Quick Filters">
          <CommandItem onSelect={() => runCommand(() => router.push("/inbox?filter=unread"))}>
            <Mail />
            <span>Unread Emails</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/inbox?filter=starred"))}>
            <Star />
            <span>Starred Emails</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => router.push("/inbox?filter=archived"))}>
            <Archive />
            <span>Archived Emails</span>
          </CommandItem>
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Actions">
          <CommandItem onSelect={() => runCommand(() => router.push("/digests"))}>
            <Sparkles />
            <span>Generate Digest</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(handleExport)}>
            <Download />
            <span>Export Data</span>
          </CommandItem>
          <CommandItem
            onSelect={() =>
              runCommand(() => setTheme(theme === "dark" ? "light" : "dark"))
            }
          >
            {mounted && theme === "dark" ? <Sun /> : <Moon />}
            <span>Toggle {mounted && theme === "dark" ? "Light" : "Dark"} Mode</span>
          </CommandItem>
        </CommandGroup>
      </CommandList>

      <div className="border-t px-3 py-2 text-xs text-muted-foreground flex items-center gap-4">
        <span className="flex items-center gap-1">
          <kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">↑↓</kbd>
          navigate
        </span>
        <span className="flex items-center gap-1">
          <kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">Enter</kbd>
          select
        </span>
        <span className="flex items-center gap-1">
          <kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">Esc</kbd>
          close
        </span>
        <CommandShortcut className="ml-auto">
          <kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">⌘K</kbd>
        </CommandShortcut>
      </div>
    </CommandDialog>
  );
}
