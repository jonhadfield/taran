"use client";

import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { apiPatch } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Archive, Star, StarOff } from "lucide-react";

interface Props {
  emailId: string;
  isStarred: boolean;
  isArchived: boolean;
}

export function InboxRowActions({ emailId, isStarred, isArchived }: Props) {
  const router = useRouter();

  const toggleStar = async (e: React.MouseEvent) => {
    e.preventDefault();
    try {
      await apiPatch(`emails/${emailId}`, { IsStarred: !isStarred });
      router.refresh();
    } catch {
      toast.error("Failed to update email");
    }
  };

  const toggleArchive = async (e: React.MouseEvent) => {
    e.preventDefault();
    try {
      await apiPatch(`emails/${emailId}`, { IsArchived: !isArchived });
      toast.success(isArchived ? "Unarchived" : "Archived");
      router.refresh();
    } catch {
      toast.error("Failed to update email");
    }
  };

  return (
    <div className="flex items-center gap-1 sm:opacity-0 sm:group-hover/row:opacity-100 sm:focus-within:opacity-100 transition-opacity">
      <Button
        variant="ghost"
        size="icon"
        className="size-8"
        onClick={toggleStar}
        aria-label={isStarred ? "Unstar email" : "Star email"}
      >
        {isStarred ? (
          <Star className="size-4 fill-yellow-400 text-yellow-400" />
        ) : (
          <StarOff className="size-4 text-muted-foreground" />
        )}
      </Button>
      <Button
        variant="ghost"
        size="icon"
        className="size-8"
        onClick={toggleArchive}
        aria-label={isArchived ? "Unarchive email" : "Archive email"}
      >
        <Archive
          className={`size-4 ${isArchived ? "text-primary" : "text-muted-foreground"}`}
        />
      </Button>
    </div>
  );
}
