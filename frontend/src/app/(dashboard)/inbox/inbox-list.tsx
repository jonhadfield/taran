"use client";

import { usePolling } from "@/hooks/use-polling";
import type { Email, ListResponse } from "@/types/api";
import { Inbox } from "lucide-react";
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

interface InboxListProps {
  initialEmails: Email[];
  initialTotal: number;
  filter: string;
  queryString: string;
}

export function InboxList({
  initialEmails,
  initialTotal,
  filter,
  queryString,
}: InboxListProps) {
  const res = usePolling<ListResponse<Email>>(
    `emails?${queryString}`,
    { data: initialEmails, total: initialTotal }
  );

  const emails = res.data || [];
  const total = res.total;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-balance">Inbox</h1>
        <span className="text-sm text-muted-foreground">{total} emails</span>
      </div>

      <InboxFilters />

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
                  className={`flex size-9 shrink-0 items-center justify-center rounded-full text-sm font-medium text-white ${getAvatarColor(senderName)}`}
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
                  <p className="text-sm text-muted-foreground truncate">
                    {email.Subject}
                  </p>
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
