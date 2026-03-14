"use client";

import { useState, useCallback, useRef, useEffect, useMemo } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { apiPatch } from "@/lib/api";
import { usePolling } from "@/hooks/use-polling";
import type { Email, ListResponse } from "@/types/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Inbox, Search, X, SlidersHorizontal, Paperclip, Calendar } from "lucide-react";
import { CopyEmailAddress } from "@/components/copy-email-address";
import Link from "next/link";
import { InboxFilters } from "./inbox-filters";
import { InboxRowActions } from "./inbox-row-actions";
import { BulkActionBar } from "./bulk-action-bar";

const avatarColors = [
  "bg-blue-500",
  "bg-emerald-500",
  "bg-amber-500",
  "bg-rose-500",
  "bg-purple-500",
  "bg-cyan-500",
  "bg-indigo-500",
  "bg-pink-500",
];

function getAvatarColor(name: string) {
  const charCode = name.charCodeAt(0) || 0;
  return avatarColors[charCode % avatarColors.length];
}

const PAGE_SIZE = 50;

interface SearchFilters {
  hasAttachment: boolean;
  since: string;
  before: string;
}

export function buildQueryString(
  filter: string,
  search: string,
  topic: string,
  limit: number,
  category?: string,
  searchFilters?: SearchFilters,
) {
  const params = [`limit=${limit}`];
  if (filter === "unread") params.push("is_read=false");
  if (filter === "starred") params.push("is_starred=true");
  if (filter === "archived") params.push("is_archived=true");
  if (search) params.push(`search=${encodeURIComponent(search)}`);
  if (topic) params.push(`topic=${encodeURIComponent(topic)}`);
  if (category) params.push(`category=${encodeURIComponent(category)}`);
  if (searchFilters?.hasAttachment) params.push("has_attachment=true");
  if (searchFilters?.since) params.push(`since=${searchFilters.since}`);
  if (searchFilters?.before) params.push(`before=${searchFilters.before}`);
  return params.join("&");
}

const CATEGORY_LABELS: Record<string, string> = {
  newsletter: "Newsletters",
  personal: "Personal",
  transactional: "Transactional",
  marketing: "Marketing",
  notification: "Notifications",
  other: "Other",
};

interface InboxListProps {
  initialEmails: Email[];
  initialTotal: number;
  filter: string;
  queryString: string;
  initialTopics: string[];
  emailAddress: string;
}

export function InboxList({
  initialEmails,
  initialTotal,
  filter: initialFilter,
  initialTopics,
  emailAddress,
}: InboxListProps) {
  const [filter, setFilter] = useState(initialFilter);
  const [searchInput, setSearchInput] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [activeTopic, setActiveTopic] = useState("");
  const [activeCategory, setActiveCategory] = useState("");
  const [limit, setLimit] = useState(PAGE_SIZE);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [showFilters, setShowFilters] = useState(false);
  const [searchFilters, setSearchFilters] = useState<SearchFilters>({
    hasAttachment: false,
    since: "",
    before: "",
  });
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const router = useRouter();
  const topics = initialTopics;
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const activeFilterCount = [
    searchFilters.hasAttachment,
    searchFilters.since,
    searchFilters.before,
  ].filter(Boolean).length;
  const queryString = buildQueryString(filter, debouncedSearch, activeTopic, limit, activeCategory, searchFilters);

  useEffect(() => {
    if (debounceTimer.current) clearTimeout(debounceTimer.current);
    debounceTimer.current = setTimeout(() => {
      setDebouncedSearch(searchInput);
    }, 300);
    return () => {
      if (debounceTimer.current) clearTimeout(debounceTimer.current);
    };
  }, [searchInput]);

  const { data: res } = usePolling<ListResponse<Email>>(
    `emails?${queryString}`,
    { data: initialEmails, total: initialTotal }
  );

  const emails = useMemo(() => res.data || [], [res.data]);
  const total = res.total;

  const handleFilterChange = useCallback((value: string) => {
    setFilter(value);
    setLimit(PAGE_SIZE);
    setSelectedIds(new Set());
    const url = value === "all" ? "/inbox" : `/inbox?filter=${value}`;
    window.history.replaceState(null, "", url);
  }, []);

  const toggleSelect = useCallback((id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const toggleSelectAll = useCallback(() => {
    setSelectedIds((prev) => {
      if (prev.size === emails.length) return new Set();
      return new Set(emails.map((e) => e.ID));
    });
  }, [emails]);

  // Derive a clamped focused index without storing it separately
  const activeFocusIndex = emails.length === 0 ? -1 : Math.min(focusedIndex, emails.length - 1);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const tag = target.tagName;
      // Don't intercept when typing in inputs (except for Escape)
      if ((tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") && e.key !== "Escape") {
        return;
      }

      switch (e.key) {
        case "j": {
          e.preventDefault();
          setFocusedIndex((prev) => Math.min(prev + 1, emails.length - 1));
          break;
        }
        case "k": {
          e.preventDefault();
          setFocusedIndex((prev) => Math.max(prev - 1, 0));
          break;
        }
        case "x": {
          e.preventDefault();
          if (focusedIndex >= 0 && focusedIndex < emails.length) {
            toggleSelect(emails[focusedIndex].ID);
          }
          break;
        }
        case "s": {
          e.preventDefault();
          if (focusedIndex >= 0 && focusedIndex < emails.length) {
            const email = emails[focusedIndex];
            apiPatch(`emails/${email.ID}`, { IsStarred: !email.IsStarred })
              .then(() => router.refresh())
              .catch(() => toast.error("Failed to update email"));
          }
          break;
        }
        case "e": {
          e.preventDefault();
          if (focusedIndex >= 0 && focusedIndex < emails.length) {
            const email = emails[focusedIndex];
            apiPatch(`emails/${email.ID}`, { IsArchived: !email.IsArchived })
              .then(() => {
                toast.success(email.IsArchived ? "Unarchived" : "Archived");
                router.refresh();
              })
              .catch(() => toast.error("Failed to update email"));
          }
          break;
        }
        case "/": {
          e.preventDefault();
          searchInputRef.current?.focus();
          break;
        }
        case "Escape": {
          e.preventDefault();
          if (tag === "INPUT" || tag === "TEXTAREA") {
            (target as HTMLInputElement).blur();
          }
          setFocusedIndex(-1);
          setSelectedIds(new Set());
          break;
        }
        case "Enter":
        case "o": {
          e.preventDefault();
          if (focusedIndex >= 0 && focusedIndex < emails.length) {
            router.push(`/inbox/${emails[focusedIndex].ID}`);
          }
          break;
        }
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [emails, focusedIndex, toggleSelect, router]);

  // Scroll focused row into view
  useEffect(() => {
    if (activeFocusIndex >= 0) {
      const row = document.querySelector(`[data-inbox-row="${activeFocusIndex}"]`);
      row?.scrollIntoView({ block: "nearest" });
    }
  }, [activeFocusIndex]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-balance">Inbox</h1>
        <span className="text-sm text-muted-foreground">{total} emails</span>
      </div>

      {/* Search */}
      <div className="space-y-2">
        <div className="flex gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
            <input
              ref={searchInputRef}
              type="text"
              placeholder="Search emails... (press /)"
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

      <InboxFilters value={filter} onChange={handleFilterChange} />

      {/* Category filter */}
      <div className="flex flex-wrap gap-1.5">
        <Badge
          variant={activeCategory === "" ? "default" : "outline"}
          className="cursor-pointer"
          role="button"
          aria-pressed={activeCategory === ""}
          tabIndex={0}
          onClick={() => setActiveCategory("")}
          onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveCategory(""); } }}
        >
          All types
        </Badge>
        {Object.entries(CATEGORY_LABELS).map(([value, label]) => (
          <Badge
            key={value}
            variant={activeCategory === value ? "default" : "outline"}
            className="cursor-pointer"
            role="button"
            aria-pressed={activeCategory === value}
            tabIndex={0}
            onClick={() => setActiveCategory(activeCategory === value ? "" : value)}
            onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveCategory(activeCategory === value ? "" : value); } }}
          >
            {label}
          </Badge>
        ))}
      </div>

      {topics.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          <Badge
            variant={activeTopic === "" ? "default" : "outline"}
            className="cursor-pointer"
            role="button"
            aria-pressed={activeTopic === ""}
            tabIndex={0}
            onClick={() => setActiveTopic("")}
            onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveTopic(""); } }}
          >
            All
          </Badge>
          {topics.map((topic) => (
            <Badge
              key={topic}
              variant={activeTopic === topic ? "default" : "outline"}
              className="cursor-pointer"
              role="button"
              aria-pressed={activeTopic === topic}
              tabIndex={0}
              onClick={() => setActiveTopic(activeTopic === topic ? "" : topic)}
              onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveTopic(activeTopic === topic ? "" : topic); } }}
            >
              {topic}
            </Badge>
          ))}
        </div>
      )}

      <BulkActionBar
        selectedIds={selectedIds}
        onClear={() => setSelectedIds(new Set())}
      />

      {emails.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Inbox className="size-12 text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium">
            {filter === "all" ? "No emails yet" : `No ${filter} emails`}
          </h3>
          <p className="text-sm text-muted-foreground mt-2">
            {filter === "all"
              ? "Forward your newsletters to your inbox to get started:"
              : "Nothing to show for this filter."}
          </p>
          {filter === "all" && emailAddress && (
            <div className="mt-3">
              <CopyEmailAddress emailAddress={emailAddress} />
            </div>
          )}
        </div>
      ) : (
        <>
          <div className="divide-y rounded-lg border">
            {/* Select all header */}
            {emails.length > 0 && (
              <div className="flex items-center gap-3 px-3.5 py-2 bg-muted/30">
                <input
                  type="checkbox"
                  className="size-4 rounded border-input accent-primary cursor-pointer"
                  checked={selectedIds.size === emails.length && emails.length > 0}
                  ref={(el) => {
                    if (el) el.indeterminate = selectedIds.size > 0 && selectedIds.size < emails.length;
                  }}
                  onChange={toggleSelectAll}
                  aria-label="Select all emails"
                />
                <span className="text-xs text-muted-foreground">
                  {selectedIds.size > 0
                    ? `${selectedIds.size} of ${emails.length} selected`
                    : "Select all"}
                </span>
              </div>
            )}
            {emails.map((email, index) => {
              const senderName = email.FromName || email.FromAddress;
              const initial = senderName.charAt(0).toUpperCase();
              const isSelected = selectedIds.has(email.ID);
              const isFocused = activeFocusIndex === index;

              return (
                <div
                  key={email.ID}
                  data-inbox-row={index}
                  className={`group/row flex items-center gap-3 p-3.5 transition-colors hover:bg-accent/50 ${isSelected ? "bg-accent/30" : ""} ${isFocused ? "ring-2 ring-inset ring-primary/50 bg-accent/40" : ""}`}
                >
                  {/* Checkbox */}
                  <input
                    type="checkbox"
                    className="size-4 shrink-0 rounded border-input accent-primary cursor-pointer"
                    checked={isSelected}
                    onChange={() => toggleSelect(email.ID)}
                    onClick={(e) => e.stopPropagation()}
                    aria-label={`Select ${email.Subject}`}
                  />

                  <Link
                    href={`/inbox/${email.ID}`}
                    className="flex flex-1 items-center gap-3 min-w-0"
                  >
                    {/* Unread dot */}
                    <div className="w-2 shrink-0">
                      {!email.IsRead && (
                        <span className="block size-2 rounded-full bg-blue-500" />
                      )}
                    </div>

                    {/* Avatar */}
                    <div
                      className={`hidden sm:flex size-9 shrink-0 items-center justify-center rounded-full text-sm font-medium text-white ${getAvatarColor(senderName)}`}
                    >
                      {initial}
                    </div>

                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-2">
                        <span
                          className={`text-sm truncate ${!email.IsRead ? "font-semibold" : ""}`}
                        >
                          {senderName}
                        </span>
                        <span className="text-xs text-muted-foreground whitespace-nowrap">
                          {new Date(email.ReceivedAt).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
                        </span>
                      </div>
                      <div className="flex items-center gap-1.5">
                        <p className="text-sm text-muted-foreground truncate">
                          {email.Subject}
                        </p>
                        {email.Status === "skipped" && (
                          <Badge variant="outline" className="shrink-0 text-xs text-muted-foreground">
                            Skipped
                          </Badge>
                        )}
                        {email.Status === "failed" && (
                          <Badge variant="outline" className="shrink-0 text-xs text-destructive">
                            Failed
                          </Badge>
                        )}
                      </div>
                    </div>
                  </Link>

                  {/* Actions - visible on hover */}
                  <InboxRowActions
                    emailId={email.ID}
                    isStarred={email.IsStarred}
                    isArchived={email.IsArchived}
                  />
                </div>
              );
            })}
          </div>
          {emails.length < total && (
            <div className="flex justify-center pt-2">
              <Button
                variant="outline"
                onClick={() => setLimit((prev) => prev + PAGE_SIZE)}
              >
                Load more ({total - emails.length} remaining)
              </Button>
            </div>
          )}
          <div className="flex justify-center pt-1">
            <p className="text-xs text-muted-foreground">
              <kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">j</kbd>/<kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">k</kbd> navigate
              {" "}<kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">x</kbd> select
              {" "}<kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">e</kbd> archive
              {" "}<kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">s</kbd> star
              {" "}<kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">/</kbd> search
              {" "}<kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">Enter</kbd> open
              {" "}<kbd className="px-1 py-0.5 rounded border bg-muted text-[10px]">Esc</kbd> clear
            </p>
          </div>
        </>
      )}
    </div>
  );
}
