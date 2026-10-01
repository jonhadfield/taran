import { useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { apiPatch } from "@/lib/api";
import type { Email } from "@/types/api";

interface UseInboxKeyboardShortcutsParams {
  emails: Email[];
  focusedIndex: number;
  setFocusedIndex: React.Dispatch<React.SetStateAction<number>>;
  toggleSelect: (id: string) => void;
  setSelectedIds: React.Dispatch<React.SetStateAction<Set<string>>>;
  setPreviewId: React.Dispatch<React.SetStateAction<string | null>>;
  searchInputRef: React.RefObject<HTMLInputElement | null>;
  isDesktop: boolean;
  refresh: () => void;
}

export function useInboxKeyboardShortcuts({
  emails,
  focusedIndex,
  setFocusedIndex,
  toggleSelect,
  setSelectedIds,
  setPreviewId,
  searchInputRef,
  isDesktop,
  refresh,
}: UseInboxKeyboardShortcutsParams) {
  const router = useRouter();

  // Keep changing values in refs so the document keydown listener stays
  // attached across email list polls / focus changes. Re-registering on
  // every emails update tears the listener down briefly and races the
  // "/" shortcut (especially under parallel E2E workers).
  const emailsRef = useRef(emails);
  const focusedIndexRef = useRef(focusedIndex);
  const isDesktopRef = useRef(isDesktop);
  const toggleSelectRef = useRef(toggleSelect);
  const refreshRef = useRef(refresh);
  const setFocusedIndexRef = useRef(setFocusedIndex);
  const setSelectedIdsRef = useRef(setSelectedIds);
  const setPreviewIdRef = useRef(setPreviewId);

  useEffect(() => {
    emailsRef.current = emails;
    focusedIndexRef.current = focusedIndex;
    isDesktopRef.current = isDesktop;
    toggleSelectRef.current = toggleSelect;
    refreshRef.current = refresh;
    setFocusedIndexRef.current = setFocusedIndex;
    setSelectedIdsRef.current = setSelectedIds;
    setPreviewIdRef.current = setPreviewId;
  });

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const tag = target.tagName;
      // Don't intercept when typing in inputs (except for Escape)
      if ((tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") && e.key !== "Escape") {
        return;
      }

      const emails = emailsRef.current;
      const focusedIndex = focusedIndexRef.current;
      const isDesktop = isDesktopRef.current;

      switch (e.key) {
        case "j": {
          e.preventDefault();
          setFocusedIndexRef.current((prev) => {
            const next = Math.min(prev + 1, emails.length - 1);
            if (isDesktop && next >= 0 && next < emails.length) {
              setPreviewIdRef.current(emails[next].ID);
            }
            return next;
          });
          break;
        }
        case "k": {
          e.preventDefault();
          setFocusedIndexRef.current((prev) => {
            const next = Math.max(prev - 1, 0);
            if (isDesktop && next >= 0 && next < emails.length) {
              setPreviewIdRef.current(emails[next].ID);
            }
            return next;
          });
          break;
        }
        case "x": {
          e.preventDefault();
          if (focusedIndex >= 0 && focusedIndex < emails.length) {
            toggleSelectRef.current(emails[focusedIndex].ID);
          }
          break;
        }
        case "s": {
          e.preventDefault();
          if (focusedIndex >= 0 && focusedIndex < emails.length) {
            const email = emails[focusedIndex];
            apiPatch(`emails/${email.ID}`, { IsStarred: !email.IsStarred })
              .then(() => refreshRef.current())
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
                refreshRef.current();
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
          setFocusedIndexRef.current(-1);
          setSelectedIdsRef.current(new Set());
          setPreviewIdRef.current(null);
          break;
        }
        case "Enter":
        case "o": {
          e.preventDefault();
          if (focusedIndex >= 0 && focusedIndex < emails.length) {
            if (isDesktop) {
              setPreviewIdRef.current(emails[focusedIndex].ID);
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
  }, [router, searchInputRef]);
}
