"use client";

import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

const shortcuts = [
  { category: "Navigation", items: [
    { keys: ["j"], description: "Next email" },
    { keys: ["k"], description: "Previous email" },
    { keys: ["Enter"], description: "Open email / preview" },
    { keys: ["/"], description: "Focus search" },
    { keys: ["Esc"], description: "Clear selection / close" },
    { keys: ["⌘", "K"], description: "Command palette" },
  ]},
  { category: "Actions", items: [
    { keys: ["x"], description: "Toggle select" },
    { keys: ["s"], description: "Toggle star" },
    { keys: ["e"], description: "Toggle archive" },
  ]},
  { category: "Help", items: [
    { keys: ["?"], description: "Show this help" },
  ]},
];

export function KeyboardHelp() {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const tag = target.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;

      if (e.key === "?") {
        e.preventDefault();
        setOpen(true);
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Keyboard Shortcuts</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          {shortcuts.map((group) => (
            <div key={group.category}>
              <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">
                {group.category}
              </h3>
              <div className="space-y-1.5">
                {group.items.map((item) => (
                  <div key={item.description} className="flex items-center justify-between text-sm">
                    <span>{item.description}</span>
                    <div className="flex items-center gap-0.5">
                      {item.keys.map((key, i) => (
                        <span key={i}>
                          {i > 0 && <span className="text-muted-foreground mx-0.5">+</span>}
                          <kbd className="px-1.5 py-0.5 rounded border bg-muted text-xs font-mono">
                            {key}
                          </kbd>
                        </span>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
