"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { toast } from "sonner";
import { apiGet, apiPut, apiDelete } from "@/lib/api";
import type { Label } from "@/types/api";
import { Tag, Plus, X } from "lucide-react";
import { labelColorClass } from "@/lib/constants";
import { useClickOutside } from "@/hooks/use-click-outside";

export function EmailLabels({ emailId }: { emailId: string }) {
  const [assignedLabels, setAssignedLabels] = useState<Label[]>([]);
  const [allLabels, setAllLabels] = useState<Label[]>([]);
  const [showMenu, setShowMenu] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    apiGet<Label[]>(`emails/${emailId}/labels`)
      .then((data) => setAssignedLabels(data || []))
      .catch(() => {});
    apiGet<Label[]>("labels")
      .then((data) => setAllLabels(data || []))
      .catch(() => {});
  }, [emailId]);

  const closeMenu = useCallback(() => setShowMenu(false), []);
  useClickOutside(menuRef, closeMenu);

  const addLabel = async (label: Label) => {
    try {
      await apiPut(`emails/${emailId}/labels/${label.ID}`, {});
      setAssignedLabels((prev) => [...prev, label]);
      setShowMenu(false);
    } catch {
      toast.error("Failed to add label");
    }
  };

  const removeLabel = async (labelId: string) => {
    try {
      await apiDelete(`emails/${emailId}/labels/${labelId}`);
      setAssignedLabels((prev) => prev.filter((l) => l.ID !== labelId));
    } catch {
      toast.error("Failed to remove label");
    }
  };

  const unassignedLabels = allLabels.filter(
    (l) => !assignedLabels.some((a) => a.ID === l.ID)
  );

  if (allLabels.length === 0) return null;

  return (
    <div className="flex items-center gap-2 flex-wrap">
      <Tag className="size-3.5 text-muted-foreground shrink-0" />
      {assignedLabels.map((label) => (
        <span
          key={label.ID}
          className="inline-flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-medium"
        >
          {label.Color && (
            <span className={`size-2 rounded-full ${labelColorClass(label.Color)}`} />
          )}
          {label.Name}
          <button
            onClick={() => removeLabel(label.ID)}
            className="text-muted-foreground hover:text-foreground ml-0.5"
            aria-label={`Remove ${label.Name} label`}
          >
            <X className="size-3" />
          </button>
        </span>
      ))}
      {unassignedLabels.length > 0 && (
        <div className="relative" ref={menuRef}>
          <button
            onClick={() => setShowMenu(!showMenu)}
            className="inline-flex items-center gap-1 rounded-full border border-dashed px-2 py-0.5 text-xs text-muted-foreground hover:text-foreground hover:border-foreground/30 transition-colors"
          >
            <Plus className="size-3" />
            Add
          </button>
          {showMenu && (
            <div className="absolute top-full left-0 mt-1 z-50 min-w-[140px] rounded-md border bg-popover p-1 shadow-md">
              {unassignedLabels.map((label) => (
                <button
                  key={label.ID}
                  onClick={() => addLabel(label)}
                  className="flex items-center gap-2 w-full rounded-sm px-2 py-1.5 text-sm hover:bg-accent transition-colors text-left"
                >
                  {label.Color && (
                    <span className={`size-2 rounded-full shrink-0 ${labelColorClass(label.Color)}`} />
                  )}
                  {label.Name}
                </button>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
