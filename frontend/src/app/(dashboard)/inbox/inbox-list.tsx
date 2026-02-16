"use client";

import { useState, useCallback, useRef, useEffect } from "react";
import { usePolling } from "@/hooks/use-polling";
import { apiGet } from "@/lib/api";
import type { Email, ListResponse } from "@/types/api";
import { Badge } from "@/components/ui/badge";
import { Inbox, Search, X } from "lucide-react";
import { APP_NAME } from "@/lib/config";
import Link from "next/link";
import { InboxFilters } from "./inbox-filters";
import { InboxRowActions } from "./inbox-row-actions";

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

function buildQueryString(filter: string, search: string, topic: string) {
  const params = ["limit=50"];
  if (filter === "unread") params.push("is_read=false");
  if (filter === "starred") params.push("is_starred=true");
  if (filter === "archived") params.push("is_archived=true");
  if (search) params.push(`search=${encodeURIComponent(search)}`);
  if (topic) params.push(`topic=${encodeURIComponent(topic)}`);
  return params.join("&");
}

interface InboxListProps {
  initialEmails: Email[];
  initialTotal: number;
  filter: string;
  queryString: string;
}

export function InboxList({
  initialEmails,
  initialTotal,
  filter: initialFilter,
}: InboxListProps) {
  const [filter, setFilter] = useState(initialFilter);
  const [searchInput, setSearchInput] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [activeTopic, setActiveTopic] = useState("");
  const [topics, setTopics] = useState<string[]>([]);
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const queryString = buildQueryString(filter, debouncedSearch, activeTopic);

  useEffect(() => {
    apiGet<string[]>("topics")
      .then(setTopics)
      .catch(() => {});
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

  const res = usePolling<ListResponse<Email>>(
    `emails?${queryString}`,
    { data: initialEmails, total: initialTotal }
  );

  const emails = res.data || [];
  const total = res.total;

  const handleFilterChange = useCallback((value: string) => {
    setFilter(value);
    const url = value === "all" ? "/inbox" : `/inbox?filter=${value}`;
    window.history.replaceState(null, "", url);
  }, []);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-balance">Inbox</h1>
        <span className="text-sm text-muted-foreground">{total} emails</span>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
        <input
          type="text"
          placeholder="Search by subject or sender..."
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          className="flex h-9 w-full rounded-md border border-input bg-transparent pl-9 pr-8 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
        {searchInput && (
          <button
            onClick={() => setSearchInput("")}
            className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
          >
            <X className="size-4" />
          </button>
        )}
      </div>

      <InboxFilters value={filter} onChange={handleFilterChange} />

      {topics.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          <Badge
            variant={activeTopic === "" ? "default" : "outline"}
            className="cursor-pointer"
            onClick={() => setActiveTopic("")}
          >
            All
          </Badge>
          {topics.map((topic) => (
            <Badge
              key={topic}
              variant={activeTopic === topic ? "default" : "outline"}
              className="cursor-pointer"
              onClick={() => setActiveTopic(activeTopic === topic ? "" : topic)}
            >
              {topic}
            </Badge>
          ))}
        </div>
      )}

      {emails.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Inbox className="size-12 text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium">
            {filter === "all" ? "No emails yet" : `No ${filter} emails`}
          </h3>
          <p className="text-sm text-muted-foreground mt-1">
            {filter === "all"
              ? `Forward your newsletters to your ${APP_NAME} inbox to get started.`
              : "Nothing to show for this filter."}
          </p>
        </div>
      ) : (
        <div className="divide-y rounded-lg border">
          {emails.map((email) => {
            const senderName = email.FromName || email.FromAddress;
            const initial = senderName.charAt(0).toUpperCase();

            return (
              <Link
                key={email.ID}
                href={`/inbox/${email.ID}`}
                className="group/row flex items-center gap-3 p-3.5 transition-colors hover:bg-accent/50"
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
                      {new Date(email.ReceivedAt).toLocaleDateString()}
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

                {/* Actions - visible on hover */}
                <InboxRowActions
                  emailId={email.ID}
                  isStarred={email.IsStarred}
                  isArchived={email.IsArchived}
                />
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}
