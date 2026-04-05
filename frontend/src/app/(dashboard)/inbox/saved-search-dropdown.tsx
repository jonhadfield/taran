"use client";

import { useState, useRef, useCallback } from "react";
import { toast } from "sonner";
import { apiPost, apiDelete } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { SavedSearch } from "@/types/api";
import { Bookmark, BookmarkPlus, Trash2 } from "lucide-react";
import { useClickOutside } from "@/hooks/use-click-outside";

interface SavedSearchDropdownProps {
  savedSearches: SavedSearch[];
  setSavedSearches: React.Dispatch<React.SetStateAction<SavedSearch[]>>;
  hasActiveFilters: boolean;
  getCurrentFilters: () => SavedSearch["Filters"];
  onApply: (s: SavedSearch) => void;
}

export function SavedSearchDropdown({
  savedSearches,
  setSavedSearches,
  hasActiveFilters,
  getCurrentFilters,
  onApply,
}: SavedSearchDropdownProps) {
  const [showSavedSearches, setShowSavedSearches] = useState(false);
  const [savingSearch, setSavingSearch] = useState(false);
  const [saveSearchName, setSaveSearchName] = useState("");
  const savedSearchRef = useRef<HTMLDivElement>(null);

  const closeSavedSearches = useCallback(() => {
    setShowSavedSearches(false);
    setSavingSearch(false);
    setSaveSearchName("");
  }, []);
  useClickOutside(savedSearchRef, closeSavedSearches);

  const handleSaveSearch = useCallback(async () => {
    if (!saveSearchName.trim()) return;
    try {
      const updated = await apiPost<SavedSearch[]>("saved-searches", {
        Name: saveSearchName.trim(),
        Filters: getCurrentFilters(),
      });
      setSavedSearches(updated || []);
      setSavingSearch(false);
      setSaveSearchName("");
      toast.success("Search saved");
    } catch {
      toast.error("Failed to save search");
    }
  }, [saveSearchName, getCurrentFilters, setSavedSearches]);

  const handleDeleteSavedSearch = useCallback(async (id: string) => {
    try {
      await apiDelete(`saved-searches/${id}`);
      setSavedSearches((prev) => prev.filter((s) => s.ID !== id));
      toast.success("Saved search deleted");
    } catch {
      toast.error("Failed to delete");
    }
  }, [setSavedSearches]);

  return (
    <div className="relative" ref={savedSearchRef}>
      <button
        onClick={() => { setShowSavedSearches(!showSavedSearches); setSavingSearch(false); }}
        className={`inline-flex items-center gap-1.5 rounded-md border px-3 h-9 text-sm transition-colors ${
          showSavedSearches
            ? "bg-primary text-primary-foreground border-primary"
            : "hover:bg-accent"
        }`}
        aria-label="Saved searches"
      >
        <Bookmark className="size-3.5" />
        <span className="hidden sm:inline">Saved</span>
        {savedSearches.length > 0 && (
          <span className="inline-flex items-center justify-center size-5 rounded-full bg-background text-foreground text-xs font-medium">
            {savedSearches.length}
          </span>
        )}
      </button>
      {(showSavedSearches || savingSearch) && (
        <div className="absolute right-0 top-full mt-1 z-50 min-w-[240px] max-w-[320px] rounded-md border bg-popover p-1 shadow-md">
          {savingSearch ? (
            <div className="p-2 space-y-2">
              <p className="text-xs font-medium text-muted-foreground">Save current filters</p>
              <Input
                autoFocus
                type="text"
                placeholder="Search name..."
                value={saveSearchName}
                onChange={(e) => setSaveSearchName(e.target.value)}
                onKeyDown={(e) => { if (e.key === "Enter") handleSaveSearch(); if (e.key === "Escape") { setSavingSearch(false); setSaveSearchName(""); } }}
                className="h-8"
              />
              <div className="flex gap-1">
                <Button size="sm" className="h-7 text-xs flex-1" onClick={handleSaveSearch} disabled={!saveSearchName.trim()}>
                  Save
                </Button>
                <Button size="sm" variant="ghost" className="h-7 text-xs" onClick={() => { setSavingSearch(false); setSaveSearchName(""); }}>
                  Cancel
                </Button>
              </div>
            </div>
          ) : (
            <>
              {hasActiveFilters && (
                <button
                  onClick={() => { setSavingSearch(true); setShowSavedSearches(false); }}
                  className="flex items-center gap-2 w-full rounded-sm px-2 py-1.5 text-sm hover:bg-accent transition-colors text-left"
                >
                  <BookmarkPlus className="size-4 shrink-0 text-muted-foreground" />
                  Save current filters
                </button>
              )}
              {hasActiveFilters && savedSearches.length > 0 && (
                <div className="my-1 border-t" />
              )}
              {savedSearches.length === 0 && !hasActiveFilters && (
                <p className="px-2 py-3 text-sm text-muted-foreground text-center">
                  Apply filters, then save them here for quick access.
                </p>
              )}
              {savedSearches.map((s) => (
                <div key={s.ID} className="group/saved flex items-center gap-1 rounded-sm hover:bg-accent transition-colors">
                  <button
                    onClick={() => { onApply(s); setShowSavedSearches(false); }}
                    className="flex-1 flex items-center gap-2 px-2 py-1.5 text-sm text-left min-w-0"
                  >
                    <Bookmark className="size-3.5 shrink-0 text-muted-foreground" />
                    <span className="truncate">{s.Name}</span>
                  </button>
                  <button
                    onClick={(e) => { e.stopPropagation(); handleDeleteSavedSearch(s.ID); }}
                    className="shrink-0 p-1 mr-1 rounded-sm opacity-0 group-hover/saved:opacity-100 hover:bg-destructive/10 hover:text-destructive transition-all"
                    aria-label={`Delete ${s.Name}`}
                  >
                    <Trash2 className="size-3.5" />
                  </button>
                </div>
              ))}
            </>
          )}
        </div>
      )}
    </div>
  );
}
