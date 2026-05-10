"use client";

import { useState, useCallback, useRef, useEffect, useMemo } from "react";
import { useEmailNotifications } from "@/hooks/use-email-notifications";
import { toast } from "sonner";
import { apiGet } from "@/lib/api";
import { usePolling } from "@/hooks/use-polling";
import type { Email, Label, SavedSearch, ListResponse } from "@/types/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState, InboxIllustration, SearchIllustration } from "@/components/empty-state";
import { Tag } from "lucide-react";
import { Kbd } from "@/components/ui/kbd";
import { CopyEmailAddress } from "@/components/copy-email-address";
import Link from "next/link";
import { InboxFilters } from "./inbox-filters";
import { InboxRowActions } from "./inbox-row-actions";
import { BulkActionBar } from "./bulk-action-bar";
import { EmailPreview } from "./email-preview";
import { InboxSearchBar } from "./inbox-search-bar";
import { useMediaQuery } from "@/hooks/use-media-query";
import { useEventSource } from "@/hooks/use-event-source";
import { useInboxKeyboardShortcuts } from "@/hooks/use-inbox-keyboard-shortcuts";
import { formatShortDate, pluralize } from "@/lib/utils";
import { labelColorClass } from "@/lib/constants";
import { SenderAvatar } from "@/components/sender-avatar";
import { CATEGORY_LABELS } from "@/lib/category-constants";

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

interface InboxListProps {
  initialEmails: Email[];
  initialTotal: number;
  filter: string;
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
  const [previewId, setPreviewId] = useState<string | null>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
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
  useInboxKeyboardShortcuts({
    emails,
    focusedIndex,
    setFocusedIndex,
    toggleSelect,
    setSelectedIds,
    setPreviewId,
    searchInputRef,
    isDesktop,
    refresh,
  });

  // Scroll focused row into view
  useEffect(() => {
    if (activeFocusIndex >= 0) {
      const row = document.querySelector(`[data-inbox-row="${activeFocusIndex}"]`);
      row?.scrollIntoView({ block: "nearest" });
    }
  }, [activeFocusIndex]);

  const hasActiveFilters = !!(filter !== "all" || debouncedSearch || activeTopic || activeCategory || sort !== "newest" || activeLabel || searchFilters.hasAttachment || searchFilters.since || searchFilters.before);

  const getCurrentFilters = useCallback((): SavedSearch["Filters"] => {
    const f: SavedSearch["Filters"] = {};
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
    const url = s.Filters.filter && s.Filters.filter !== "all" ? `/inbox?filter=${s.Filters.filter}` : "/inbox";
    window.history.replaceState(null, "", url);
  }, []);

  const listContent = (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-balance">Inbox</h1>
        <span className="text-sm text-muted-foreground">
          {debouncedSearch
            ? `${total} ${pluralize(total, "result")} for "${debouncedSearch}"`
            : `${total} emails`}
        </span>
      </div>

      {/* Search */}
      <InboxSearchBar
        searchInput={searchInput}
        setSearchInput={setSearchInput}
        sort={sort}
        setSort={setSort}
        setLimit={setLimit}
        pageSize={PAGE_SIZE}
        showFilters={showFilters}
        setShowFilters={setShowFilters}
        searchFilters={searchFilters}
        setSearchFilters={setSearchFilters}
        activeFilterCount={activeFilterCount}
        debouncedSearch={debouncedSearch}
        searchInputRef={searchInputRef}
        savedSearches={savedSearches}
        setSavedSearches={setSavedSearches}
        hasActiveFilters={hasActiveFilters}
        getCurrentFilters={getCurrentFilters}
        onApplySavedSearch={applySavedSearch}
      />

      <InboxFilters value={filter} onChange={handleFilterChange} />

      {/* Category filter */}
      <div className="flex flex-wrap gap-1.5">
        <Badge
          variant={activeCategory === "" ? "default" : "outline"}
          className="cursor-pointer"
          role="button"
          aria-pressed={activeCategory === ""}
          tabIndex={0}
          onClick={() => { setActiveCategory(""); setLimit(PAGE_SIZE); }}
          onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveCategory(""); setLimit(PAGE_SIZE); } }}
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
            onClick={() => { setActiveCategory(activeCategory === value ? "" : value); setLimit(PAGE_SIZE); }}
            onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveCategory(activeCategory === value ? "" : value); setLimit(PAGE_SIZE); } }}
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
            onClick={() => { setActiveLabel(""); setLimit(PAGE_SIZE); }}
            onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveLabel(""); setLimit(PAGE_SIZE); } }}
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
              onClick={() => { setActiveLabel(activeLabel === label.ID ? "" : label.ID); setLimit(PAGE_SIZE); }}
              onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveLabel(activeLabel === label.ID ? "" : label.ID); setLimit(PAGE_SIZE); } }}
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
            onClick={() => { setActiveTopic(""); setLimit(PAGE_SIZE); }}
            onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveTopic(""); setLimit(PAGE_SIZE); } }}
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
              onClick={() => { setActiveTopic(activeTopic === topic ? "" : topic); setLimit(PAGE_SIZE); }}
              onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setActiveTopic(activeTopic === topic ? "" : topic); setLimit(PAGE_SIZE); } }}
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
        onMutate={refresh}
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
                  <SenderAvatar name={senderName} />

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
                    onMutate={refresh}
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
              <Kbd className="h-auto min-w-0 py-0.5 text-[10px]">j</Kbd>/<Kbd className="h-auto min-w-0 py-0.5 text-[10px]">k</Kbd> navigate
              {" "}<Kbd className="h-auto min-w-0 py-0.5 text-[10px]">x</Kbd> select
              {" "}<Kbd className="h-auto min-w-0 py-0.5 text-[10px]">e</Kbd> archive
              {" "}<Kbd className="h-auto min-w-0 py-0.5 text-[10px]">s</Kbd> star
              {" "}<Kbd className="h-auto min-w-0 py-0.5 text-[10px]">/</Kbd> search
              {" "}<Kbd className="h-auto min-w-0 py-0.5 text-[10px]">Enter</Kbd> open
              {" "}<Kbd className="h-auto min-w-0 py-0.5 text-[10px]">Esc</Kbd> clear
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
