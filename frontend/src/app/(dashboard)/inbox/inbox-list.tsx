"use client";

import { useState, useCallback, useRef, useEffect, useMemo } from "react";
import { useEmailNotifications } from "@/hooks/use-email-notifications";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { apiGet, apiPost, apiPatch, apiDelete } from "@/lib/api";
import { usePolling } from "@/hooks/use-polling";
import type { Email, Label, SavedSearch, ListResponse } from "@/types/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState, InboxIllustration, SearchIllustration } from "@/components/empty-state";
import { Search, X, SlidersHorizontal, Paperclip, Calendar, ArrowUpDown, Tag, Bookmark, BookmarkPlus, Trash2 } from "lucide-react";
import { CopyEmailAddress } from "@/components/copy-email-address";
import Link from "next/link";
import { InboxFilters } from "./inbox-filters";
import { InboxRowActions } from "./inbox-row-actions";
import { BulkActionBar } from "./bulk-action-bar";
import { EmailPreview } from "./email-preview";
import { useMediaQuery } from "@/hooks/use-media-query";
import { useEventSource } from "@/hooks/use-event-source";
import { formatShortDate } from "@/lib/utils";

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

const LABEL_COLORS: Record<string, string> = {
  blue: "bg-blue-500",
  green: "bg-green-500",
  red: "bg-red-500",
  purple: "bg-purple-500",
  orange: "bg-orange-500",
  yellow: "bg-yellow-500",
  pink: "bg-pink-500",
  cyan: "bg-cyan-500",
};

function labelColorClass(color: string) {
  return LABEL_COLORS[color] || "bg-gray-400";
}

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
  sort?: string,
  labelId?: string,
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
  if (sort && sort !== "newest") params.push(`sort=${sort}`);
  if (labelId) params.push(`label=${encodeURIComponent(labelId)}`);
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
  const [sort, setSort] = useState("newest");
  const [activeLabel, setActiveLabel] = useState("");
  const [labels, setLabels] = useState<Label[]>([]);
  const [savedSearches, setSavedSearches] = useState<SavedSearch[]>([]);
  const [showSavedSearches, setShowSavedSearches] = useState(false);
  const [savingSearch, setSavingSearch] = useState(false);
  const [saveSearchName, setSaveSearchName] = useState("");
  const [previewId, setPreviewId] = useState<string | null>(null);
  const savedSearchRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const router = useRouter();
  const isDesktop = useMediaQuery("(min-width: 1280px)");
  const topics = initialTopics;
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const activeFilterCount = [
    searchFilters.hasAttachment,
    searchFilters.since,
    searchFilters.before,
  ].filter(Boolean).length;
  const queryString = buildQueryString(filter, debouncedSearch, activeTopic, limit, activeCategory, searchFilters, sort, activeLabel);

  useEffect(() => {
    apiGet<Label[]>("labels").then((data) => setLabels(data || [])).catch(() => {});
    apiGet<SavedSearch[]>("saved-searches").then((data) => setSavedSearches(data || [])).catch(() => {});
  }, []);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (savedSearchRef.current && !savedSearchRef.current.contains(e.target as Node)) {
        setShowSavedSearches(false);
        setSavingSearch(false);
        setSaveSearchName("");
      }
    };
    if (showSavedSearches || savingSearch) {
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [showSavedSearches, savingSearch]);

  useEffect(() => {
    if (debounceTimer.current) clearTimeout(debounceTimer.current);
    debounceTimer.current = setTimeout(() => {
      setDebouncedSearch(searchInput);
    }, 300);
    return () => {
      if (debounceTimer.current) clearTimeout(debounceTimer.current);
    };
  }, [searchInput]);

  const { data: res, refresh } = usePolling<ListResponse<Email>>(
    `emails?${queryString}`,
    { data: initialEmails, total: initialTotal }
  );

  // Real-time updates via SSE — triggers immediate refresh when an email is processed
  const sseStatus = useEventSource(useCallback(() => {
    refresh();
  }, [refresh]));

  // Show toast on SSE reconnection
  const prevSSEStatus = useRef(sseStatus);
  useEffect(() => {
    if (prevSSEStatus.current === "reconnecting" && sseStatus === "connected") {
      toast.success("Reconnected to live updates");
    }
    prevSSEStatus.current = sseStatus;
  }, [sseStatus]);

  const emails = useMemo(() => res.data || [], [res.data]);
  const total = res.total;

  // Prompt to enable notifications on first visit
  const { permission, requestPermission, notify } = useEmailNotifications();
  const hasPrompted = useRef(false);

  useEffect(() => {
    if (hasPrompted.current) return;
    if (permission !== "default") return;
    if (typeof window === "undefined") return;
    const dismissed = localStorage.getItem("notification-prompt-dismissed");
    if (dismissed) return;
    hasPrompted.current = true;

    const timer = setTimeout(() => {
      toast("Get notified when new emails arrive", {
        description: "Enable browser notifications so you never miss a message.",
        action: {
          label: "Enable",
          onClick: () => {
            requestPermission();
          },
        },
        onDismiss: () => {
          localStorage.setItem("notification-prompt-dismissed", "1");
        },
        duration: 15000,
      });
    }, 2000);
    return () => clearTimeout(timer);
  }, [permission, requestPermission]);

  // Notify on new emails (only when "all" filter, no search — the default inbox view)
  const prevEmailIds = useRef<Set<string>>(new Set(initialEmails.map((e) => e.ID)));

  useEffect(() => {
    if (filter !== "all" || debouncedSearch) return;

    const currentIds = new Set(emails.map((e) => e.ID));
    const newEmails = emails.filter((e) => !prevEmailIds.current.has(e.ID));

    if (newEmails.length > 0 && prevEmailIds.current.size > 0) {
      notify(newEmails);
    }

    prevEmailIds.current = currentIds;
  }, [emails, filter, debouncedSearch, notify]);

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
          setFocusedIndex((prev) => {
            const next = Math.min(prev + 1, emails.length - 1);
            if (isDesktop && next >= 0 && next < emails.length) {
              setPreviewId(emails[next].ID);
            }
            return next;
          });
          break;
        }
        case "k": {
          e.preventDefault();
          setFocusedIndex((prev) => {
            const next = Math.max(prev - 1, 0);
            if (isDesktop && next >= 0 && next < emails.length) {
              setPreviewId(emails[next].ID);
            }
            return next;
          });
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
          setPreviewId(null);
          break;
        }
        case "Enter":
        case "o": {
          e.preventDefault();
          if (focusedIndex >= 0 && focusedIndex < emails.length) {
            if (isDesktop) {
              setPreviewId(emails[focusedIndex].ID);
            } else {
              router.push(`/inbox/${emails[focusedIndex].ID}`);
            }
          }
          break;
        }
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [emails, focusedIndex, toggleSelect, router, isDesktop]);

  // Scroll focused row into view
  useEffect(() => {
    if (activeFocusIndex >= 0) {
      const row = document.querySelector(`[data-inbox-row="${activeFocusIndex}"]`);
      row?.scrollIntoView({ block: "nearest" });
    }
  }, [activeFocusIndex]);

  const hasActiveFilters = filter !== "all" || debouncedSearch || activeTopic || activeCategory || sort !== "newest" || activeLabel || searchFilters.hasAttachment || searchFilters.since || searchFilters.before;

  const getCurrentFilters = useCallback(() => {
    const f: Record<string, unknown> = {};
    if (filter !== "all") f.filter = filter;
    if (debouncedSearch) f.search = debouncedSearch;
    if (activeTopic) f.topic = activeTopic;
    if (activeCategory) f.category = activeCategory;
    if (sort !== "newest") f.sort = sort;
    if (activeLabel) f.labelId = activeLabel;
    if (searchFilters.hasAttachment) f.hasAttachment = true;
    if (searchFilters.since) f.since = searchFilters.since;
    if (searchFilters.before) f.before = searchFilters.before;
    return f;
  }, [filter, debouncedSearch, activeTopic, activeCategory, sort, activeLabel, searchFilters]);

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
  }, [saveSearchName, getCurrentFilters]);

  const handleDeleteSavedSearch = useCallback(async (id: string) => {
    try {
      await apiDelete(`saved-searches/${id}`);
      setSavedSearches((prev) => prev.filter((s) => s.ID !== id));
      toast.success("Saved search deleted");
    } catch {
      toast.error("Failed to delete");
    }
  }, []);

  const applySavedSearch = useCallback((s: SavedSearch) => {
    setFilter(s.Filters.filter || "all");
    setSearchInput(s.Filters.search || "");
    setDebouncedSearch(s.Filters.search || "");
    setActiveTopic(s.Filters.topic || "");
    setActiveCategory(s.Filters.category || "");
    setSort(s.Filters.sort || "newest");
    setActiveLabel(s.Filters.labelId || "");
    setSearchFilters({
      hasAttachment: s.Filters.hasAttachment || false,
      since: s.Filters.since || "",
      before: s.Filters.before || "",
    });
    setLimit(PAGE_SIZE);
    setShowSavedSearches(false);
    const url = s.Filters.filter && s.Filters.filter !== "all" ? `/inbox?filter=${s.Filters.filter}` : "/inbox";
    window.history.replaceState(null, "", url);
  }, []);

  const listContent = (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-balance">Inbox</h1>
        <span className="text-sm text-muted-foreground">
          {debouncedSearch
            ? `${total} result${total !== 1 ? "s" : ""} for "${debouncedSearch}"`
            : `${total} emails`}
        </span>
      </div>

      {/* Search */}
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
            <ArrowUpDown className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground pointer-events-none" />
            <select
              value={sort}
              onChange={(e) => { setSort(e.target.value); setLimit(PAGE_SIZE); }}
              className="h-9 rounded-md border border-input bg-transparent pl-8 pr-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring appearance-none cursor-pointer"
              aria-label="Sort order"
            >
              <option value="newest">Newest</option>
              <option value="oldest">Oldest</option>
              {debouncedSearch && <option value="relevance">Relevance</option>}
            </select>
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
                    <input
                      autoFocus
                      type="text"
                      placeholder="Search name..."
                      value={saveSearchName}
                      onChange={(e) => setSaveSearchName(e.target.value)}
                      onKeyDown={(e) => { if (e.key === "Enter") handleSaveSearch(); if (e.key === "Escape") { setSavingSearch(false); setSaveSearchName(""); } }}
                      className="flex h-8 w-full rounded-md border border-input bg-transparent px-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
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
                          onClick={() => applySavedSearch(s)}
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

      {/* Label filter */}
      {labels.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          <Badge
            variant={activeLabel === "" ? "default" : "outline"}
            className="cursor-pointer"
            role="button"
            aria-pressed={activeLabel === ""}
            tabIndex={0}
            onClick={() => setActiveLabel("")}
            onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveLabel(""); } }}
          >
            <Tag className="size-3 mr-1" />
            All
          </Badge>
          {labels.map((label) => (
            <Badge
              key={label.ID}
              variant={activeLabel === label.ID ? "default" : "outline"}
              className="cursor-pointer"
              role="button"
              aria-pressed={activeLabel === label.ID}
              tabIndex={0}
              onClick={() => setActiveLabel(activeLabel === label.ID ? "" : label.ID)}
              onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveLabel(activeLabel === label.ID ? "" : label.ID); } }}
            >
              {label.Color && (
                <span className={`inline-block size-2 rounded-full mr-1.5 ${labelColorClass(label.Color)}`} />
              )}
              {label.Name}
              {label.EmailCount > 0 && (
                <span className="ml-1 text-muted-foreground tabular-nums">{label.EmailCount}</span>
              )}
            </Badge>
          ))}
        </div>
      )}

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
        labels={labels}
        onClear={() => setSelectedIds(new Set())}
      />

      {emails.length === 0 ? (
        <div>
          {debouncedSearch ? (
            <EmptyState
              icon={<SearchIllustration />}
              title={`No results for \u201c${debouncedSearch}\u201d`}
              description="Try a different search term or adjust your filters."
            >
              <Button
                variant="outline"
                onClick={() => { setSearchInput(""); setSearchFilters({ hasAttachment: false, since: "", before: "" }); setActiveCategory(""); setActiveTopic(""); }}
              >
                Clear search
              </Button>
            </EmptyState>
          ) : (
            <EmptyState
              icon={<InboxIllustration />}
              title={filter === "all" ? "No emails yet" : `No ${filter} emails`}
              description={filter === "all" ? "Forward your newsletters to your inbox to get started:" : "Nothing to show for this filter."}
            >
              {filter === "all" && emailAddress && (
                <CopyEmailAddress emailAddress={emailAddress} />
              )}
            </EmptyState>
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

              const isPreviewActive = isDesktop && previewId === email.ID;

              const rowInner = (
                <>
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
                        {formatShortDate(email.ReceivedAt)}
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
                </>
              );

              return (
                <div
                  key={email.ID}
                  data-inbox-row={index}
                  className={`group/row flex items-center gap-3 p-3.5 transition-colors hover:bg-accent/50 cursor-pointer ${isSelected ? "bg-accent/30" : ""} ${isFocused ? "ring-2 ring-inset ring-primary/50 bg-accent/40" : ""} ${isPreviewActive ? "border-l-2 border-l-primary bg-accent/50" : ""}`}
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

                  {isDesktop ? (
                    <button
                      type="button"
                      className="flex flex-1 items-center gap-3 min-w-0 text-left"
                      onClick={() => {
                        setPreviewId(email.ID);
                        setFocusedIndex(index);
                      }}
                    >
                      {rowInner}
                    </button>
                  ) : (
                    <Link
                      href={`/inbox/${email.ID}`}
                      className="flex flex-1 items-center gap-3 min-w-0"
                    >
                      {rowInner}
                    </Link>
                  )}

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
          <div className="hidden sm:flex justify-center pt-1">
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

  if (!isDesktop) {
    return listContent;
  }

  return (
    <div className="flex gap-4 h-[calc(100vh-10rem)]">
      <div className="w-2/5 min-w-0 overflow-y-auto">
        {listContent}
      </div>
      <div className="w-3/5 min-w-0 overflow-y-auto rounded-lg border">
        <EmailPreview
          id={previewId}
          onDeleted={() => setPreviewId(null)}
        />
      </div>
    </div>
  );
}
