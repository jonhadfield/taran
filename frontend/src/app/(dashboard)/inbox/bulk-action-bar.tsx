"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { apiPatch, apiDeleteJSON } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Archive, ArchiveRestore, Eye, EyeOff, Trash2, X } from "lucide-react";

interface BulkActionBarProps {
  selectedIds: Set<string>;
  onClear: () => void;
}

export function BulkActionBar({ selectedIds, onClear }: BulkActionBarProps) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const count = selectedIds.size;

  if (count === 0) return null;

  const ids = Array.from(selectedIds);

  const batchUpdate = async (state: Record<string, boolean>, label: string) => {
    setLoading(true);
    try {
      await apiPatch("emails/batch", { ids, state });
      toast.success(`${label} ${count} email${count > 1 ? "s" : ""}`);
      onClear();
      router.refresh();
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
      toast.success(`Deleted ${count} email${count > 1 ? "s" : ""}`);
      onClear();
      router.refresh();
    } catch {
      toast.error("Failed to delete");
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
