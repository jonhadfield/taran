"use client";

import { useEffect, useRef } from "react";
import { Search, X, SlidersHorizontal, Paperclip, Calendar, ArrowUpDown } from "lucide-react";
import type { SavedSearch } from "@/types/api";
import { NativeSelect } from "@/components/ui/native-select";
import { SavedSearchDropdown } from "./saved-search-dropdown";

interface SearchFilters {
  hasAttachment: boolean;
  since: string;
  before: string;
}

interface InboxSearchBarProps {
  searchInput: string;
  setSearchInput: (value: string) => void;
  sort: string;
  setSort: (value: string) => void;
  setLimit: React.Dispatch<React.SetStateAction<number>>;
  pageSize: number;
  showFilters: boolean;
  setShowFilters: (value: boolean) => void;
  searchFilters: SearchFilters;
  setSearchFilters: React.Dispatch<React.SetStateAction<SearchFilters>>;
  activeFilterCount: number;
  debouncedSearch: string;
  searchInputRef: React.RefObject<HTMLInputElement | null>;
  savedSearches: SavedSearch[];
  setSavedSearches: React.Dispatch<React.SetStateAction<SavedSearch[]>>;
  hasActiveFilters: boolean;
  getCurrentFilters: () => SavedSearch["Filters"];
  onApplySavedSearch: (s: SavedSearch) => void;
}

export function InboxSearchBar({
  searchInput,
  setSearchInput,
  sort,
  setSort,
  setLimit,
  pageSize,
  showFilters,
  setShowFilters,
  searchFilters,
  setSearchFilters,
  activeFilterCount,
  debouncedSearch,
  searchInputRef,
  savedSearches,
  setSavedSearches,
  hasActiveFilters,
  getCurrentFilters,
  onApplySavedSearch,
}: InboxSearchBarProps) {
  return (
    <div className="space-y-2">
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <input
            ref={searchInputRef}
            type="text"
            placeholder="Search subjects, senders, summaries... (/)"
            aria-label="Search emails"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            className="flex h-9 w-full rounded-md border border-input bg-transparent pl-9 pr-8 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          />
          {searchInput && (
            <button
              onClick={() => setSearchInput("")}
              aria-label="Clear search"
              className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              <X className="size-4" />
            </button>
          )}
        </div>
        <div className="relative">
          <ArrowUpDown className="absolute left-2.5 top-1/2 z-10 -translate-y-1/2 size-3.5 text-muted-foreground pointer-events-none" />
          <NativeSelect
            value={sort}
            onChange={(e) => { setSort(e.target.value); setLimit(pageSize); }}
            className="pl-8 cursor-pointer"
            aria-label="Sort order"
          >
            <option value="newest">Newest</option>
            <option value="oldest">Oldest</option>
            {debouncedSearch && <option value="relevance">Relevance</option>}
          </NativeSelect>
        </div>
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={`inline-flex items-center gap-1.5 rounded-md border px-3 h-9 text-sm transition-colors ${
            showFilters || activeFilterCount > 0
              ? "bg-primary text-primary-foreground border-primary"
              : "hover:bg-accent"
          }`}
          aria-label="Toggle search filters"
        >
          <SlidersHorizontal className="size-3.5" />
          <span className="hidden sm:inline">Filters</span>
          {activeFilterCount > 0 && (
            <span className="inline-flex items-center justify-center size-5 rounded-full bg-background text-foreground text-xs font-medium">
              {activeFilterCount}
            </span>
          )}
        </button>
        <SavedSearchDropdown
          savedSearches={savedSearches}
          setSavedSearches={setSavedSearches}
          hasActiveFilters={hasActiveFilters}
          getCurrentFilters={getCurrentFilters}
          onApply={onApplySavedSearch}
        />
      </div>

      {/* Expanded filters */}
      {showFilters && (
        <div className="flex flex-wrap items-end gap-2 sm:gap-3 rounded-lg border bg-muted/30 p-3">
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground flex items-center gap-1">
              <Calendar className="size-3" />
              From
            </label>
            <input
              type="date"
              value={searchFilters.since}
              onChange={(e) =>
                setSearchFilters((f) => ({ ...f, since: e.target.value }))
              }
              className="h-8 rounded-md border border-input bg-transparent px-2 text-xs shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground flex items-center gap-1">
              <Calendar className="size-3" />
              To
            </label>
            <input
              type="date"
              value={searchFilters.before}
              onChange={(e) =>
                setSearchFilters((f) => ({ ...f, before: e.target.value }))
              }
              className="h-8 rounded-md border border-input bg-transparent px-2 text-xs shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>
          <label className="inline-flex items-center gap-2 h-8 cursor-pointer">
            <input
              type="checkbox"
              checked={searchFilters.hasAttachment}
              onChange={(e) =>
                setSearchFilters((f) => ({
                  ...f,
                  hasAttachment: e.target.checked,
                }))
              }
              className="size-4 rounded border-input accent-primary"
            />
            <span className="flex items-center gap-1 text-xs font-medium">
              <Paperclip className="size-3" />
              Has attachment
            </span>
          </label>
          {activeFilterCount > 0 && (
            <button
              onClick={() =>
                setSearchFilters({
                  hasAttachment: false,
                  since: "",
                  before: "",
                })
              }
              className="h-8 rounded-md border px-2 text-xs text-muted-foreground hover:text-foreground hover:bg-accent transition-colors"
            >
              Clear filters
            </button>
          )}
        </div>
      )}
    </div>
  );
}
