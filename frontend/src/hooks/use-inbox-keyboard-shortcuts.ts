import { useEffect } from "react";
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
              .then(() => refresh())
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
                refresh();
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
  }, [emails, focusedIndex, toggleSelect, router, isDesktop, refresh, setFocusedIndex, setSelectedIds, setPreviewId, searchInputRef]);
}
