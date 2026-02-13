"use client";

import { useEffect, useRef, useState } from "react";
import { apiPatch } from "@/lib/api";
import { Button } from "@/components/ui/button";
import type { Email } from "@/types/api";
import { Archive, Star, StarOff } from "lucide-react";

export function EmailActions({ email }: { email: Email }) {
  const [starred, setStarred] = useState(email.IsStarred);
  const [archived, setArchived] = useState(email.IsArchived);
  const markedRead = useRef(false);

  useEffect(() => {
    if (!email.IsRead && !markedRead.current) {
      markedRead.current = true;
      apiPatch(`emails/${email.ID}`, { IsRead: true }).catch(() => {});
    }
  }, [email.ID, email.IsRead]);

  const toggleStar = async () => {
    await apiPatch(`emails/${email.ID}`, { IsStarred: !starred });
    setStarred(!starred);
  };

  const toggleArchive = async () => {
    await apiPatch(`emails/${email.ID}`, { IsArchived: !archived });
    setArchived(!archived);
  };

  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="sm" onClick={toggleStar}>
        {starred ? (
          <Star className="mr-1 h-4 w-4 fill-yellow-400 text-yellow-400" />
        ) : (
          <StarOff className="mr-1 h-4 w-4" />
        )}
        {starred ? "Starred" : "Star"}
      </Button>
      <Button variant="outline" size="sm" onClick={toggleArchive}>
        <Archive className="mr-1 h-4 w-4" />
        {archived ? "Unarchive" : "Archive"}
      </Button>
    </div>
  );
}
