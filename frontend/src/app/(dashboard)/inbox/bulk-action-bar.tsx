"use client";

import { useState, useRef, useCallback } from "react";
import { toast } from "sonner";
import { apiPost, apiPatch, apiDeleteJSON } from "@/lib/api";
import { pluralize } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import type { Label } from "@/types/api";
import { Archive, ArchiveRestore, Eye, EyeOff, Trash2, X, Tag } from "lucide-react";
import { labelColorClass } from "@/lib/constants";
import { useClickOutside } from "@/hooks/use-click-outside";

interface BulkActionBarProps {
  selectedIds: Set<string>;
  labels: Label[];
  onClear: () => void;
  onMutate: () => void;
}

export function BulkActionBar({ selectedIds, labels, onClear, onMutate }: BulkActionBarProps) {
  const [loading, setLoading] = useState(false);
  const [showLabelMenu, setShowLabelMenu] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const count = selectedIds.size;

  const closeLabelMenu = useCallback(() => setShowLabelMenu(false), []);
  useClickOutside(menuRef, closeLabelMenu);

  if (count === 0) return null;

  const ids = Array.from(selectedIds);

  const batchUpdate = async (state: Record<string, boolean>, label: string) => {
    setLoading(true);
    try {
      await apiPatch("emails/batch", { ids, state });
      toast.success(`${label} ${count} ${pluralize(count, "email")}`);
      onClear();
      onMutate();
    } catch {
      toast.error(`Failed to ${label.toLowerCase()}`);
    } finally {
      setLoading(false);
    }
  };

  const batchDelete = async () => {
    setLoading(true);
    try {
      await apiDeleteJSON("emails/batch", { ids });
      toast.success(`Deleted ${count} ${pluralize(count, "email")}`);
      onClear();
      onMutate();
    } catch {
      toast.error("Failed to delete");
    } finally {
      setLoading(false);
    }
  };

  const batchAddLabel = async (labelId: string, labelName: string) => {
    setLoading(true);
    try {
      await apiPost("emails/labels/batch-add", { EmailIDs: ids, LabelID: labelId });
      toast.success(`Added "${labelName}" to ${count} ${pluralize(count, "email")}`);
      setShowLabelMenu(false);
    } catch {
      toast.error("Failed to add label");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center gap-1 sm:gap-2 rounded-lg border bg-muted/50 px-2 sm:px-3 py-2 overflow-x-auto">
      <span className="text-sm font-medium mr-1 shrink-0">
        {count} selected
      </span>
      <Button
        variant="ghost"
        size="sm"
        disabled={loading}
        onClick={() => batchUpdate({ IsRead: true }, "Marked read")}
        className="shrink-0 px-2 sm:px-3"
      >
        <Eye className="size-4 sm:mr-1.5" />
        <span className="hidden sm:inline">Read</span>
      </Button>
      <Button
        variant="ghost"
        size="sm"
        disabled={loading}
        onClick={() => batchUpdate({ IsRead: false }, "Marked unread")}
        className="shrink-0 px-2 sm:px-3"
      >
        <EyeOff className="size-4 sm:mr-1.5" />
        <span className="hidden sm:inline">Unread</span>
      </Button>
      <Button
        variant="ghost"
        size="sm"
        disabled={loading}
        onClick={() => batchUpdate({ IsArchived: true }, "Archived")}
        className="shrink-0 px-2 sm:px-3"
      >
        <Archive className="size-4 sm:mr-1.5" />
        <span className="hidden sm:inline">Archive</span>
      </Button>
      <Button
        variant="ghost"
        size="sm"
        disabled={loading}
        onClick={() => batchUpdate({ IsArchived: false }, "Unarchived")}
        className="shrink-0 px-2 sm:px-3"
      >
        <ArchiveRestore className="size-4 sm:mr-1.5" />
        <span className="hidden sm:inline">Unarchive</span>
      </Button>
      {labels.length > 0 && (
        <div className="relative" ref={menuRef}>
          <Button
            variant="ghost"
            size="sm"
            disabled={loading}
            onClick={() => setShowLabelMenu(!showLabelMenu)}
            className="shrink-0 px-2 sm:px-3"
          >
            <Tag className="size-4 sm:mr-1.5" />
            <span className="hidden sm:inline">Label</span>
          </Button>
          {showLabelMenu && (
            <div className="absolute top-full left-0 mt-1 z-50 min-w-[160px] rounded-md border bg-popover p-1 shadow-md">
              {labels.map((label) => (
                <button
                  key={label.ID}
                  onClick={() => batchAddLabel(label.ID, label.Name)}
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
      <Button
        variant="ghost"
        size="sm"
        disabled={loading}
        onClick={batchDelete}
        className="text-destructive hover:text-destructive shrink-0 px-2 sm:px-3"
      >
        <Trash2 className="size-4 sm:mr-1.5" />
        <span className="hidden sm:inline">Delete</span>
      </Button>
      <div className="flex-1" />
      <Button variant="ghost" size="icon" className="size-7 shrink-0" onClick={onClear}>
        <X className="size-4" />
      </Button>
    </div>
  );
}
