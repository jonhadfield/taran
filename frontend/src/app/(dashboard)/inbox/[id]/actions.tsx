"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { apiPatch, apiDelete } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { Email } from "@/types/api";
import { Archive, Star, StarOff, Trash2, Loader2 } from "lucide-react";

export function EmailActions({ email }: { email: Email }) {
  const router = useRouter();
  const [starred, setStarred] = useState(email.IsStarred);
  const [archived, setArchived] = useState(email.IsArchived);
  const [showDelete, setShowDelete] = useState(false);
  const [deleting, setDeleting] = useState(false);
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

  const handleDelete = async () => {
    setDeleting(true);
    try {
      await apiDelete(`emails/${email.ID}`);
      router.push("/inbox");
    } catch {
      setDeleting(false);
      setShowDelete(false);
    }
  };

  return (
    <>
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
        <Button variant="outline" size="sm" onClick={() => setShowDelete(true)}>
          <Trash2 className="mr-1 h-4 w-4" />
          Delete
        </Button>
      </div>

      <Dialog open={showDelete} onOpenChange={setShowDelete}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Email</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this email? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setShowDelete(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting && <Loader2 className="mr-1 h-4 w-4 animate-spin" />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
