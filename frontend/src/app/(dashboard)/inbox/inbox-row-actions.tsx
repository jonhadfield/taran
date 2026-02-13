"use client";

import { useRouter } from "next/navigation";
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
    await apiPatch(`emails/${emailId}`, { IsStarred: !isStarred });
    router.refresh();
  };

  const toggleArchive = async (e: React.MouseEvent) => {
    e.preventDefault();
    await apiPatch(`emails/${emailId}`, { IsArchived: !isArchived });
    router.refresh();
  };

  return (
    <div className="flex items-center gap-1 opacity-0 group-hover/row:opacity-100 transition-opacity">
      <Button variant="ghost" size="icon" className="size-8" onClick={toggleStar}>
        {isStarred ? (
          <Star className="size-4 fill-yellow-400 text-yellow-400" />
        ) : (
          <StarOff className="size-4 text-muted-foreground" />
        )}
      </Button>
      <Button variant="ghost" size="icon" className="size-8" onClick={toggleArchive}>
        <Archive
          className={`size-4 ${isArchived ? "text-primary" : "text-muted-foreground"}`}
        />
      </Button>
    </div>
  );
}
