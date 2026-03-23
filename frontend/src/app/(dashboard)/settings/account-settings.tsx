"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Loader2, Trash2 } from "lucide-react";
import type { EmailAccount } from "@/types/api";
import { CopyButton } from "./copy-button";
import { UsernameForm } from "@/components/username-form";

interface AccountSettingsProps {
  accounts: EmailAccount[];
  loading: boolean;
  error: string;
  deleteTarget: EmailAccount | null;
  deleting: boolean;
  onFetchAccounts: () => void;
  onSetDeleteTarget: (account: EmailAccount | null) => void;
  onDelete: () => void;
}

export function AccountSettings({
  accounts,
  loading,
  error,
  deleteTarget,
  deleting,
  onFetchAccounts,
  onSetDeleteTarget,
  onDelete,
}: AccountSettingsProps) {
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Email Accounts</CardTitle>
          <CardDescription>
            Your managed inboxes for receiving newsletters
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {loading && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              Loading accounts...
            </div>
          )}

          {error && <p className="text-sm text-destructive">{error}</p>}

          {!loading && accounts.length === 0 && !error && (
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground">
                You haven&apos;t set up an inbox yet. Choose a username to get started.
              </p>
              <UsernameForm onSuccess={onFetchAccounts} />
            </div>
          )}

          {accounts.map((account) => (
            <div
              key={account.ID}
              className="flex items-center justify-between rounded-lg border p-4"
            >
              <div>
                <p className="font-medium">
                  {account.DisplayName || "My Inbox"}
                </p>
                <p className="text-sm text-muted-foreground">
                  {account.EmailAddress}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <CopyButton text={account.EmailAddress} />
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onSetDeleteTarget(account)}
                >
                  <Trash2 className="h-4 w-4 text-muted-foreground" />
                </Button>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <Dialog open={!!deleteTarget} onOpenChange={() => onSetDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Inbox</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete{" "}
              <span className="font-medium">
                {deleteTarget?.EmailAddress}
              </span>
              ? This will permanently delete all emails, extractions, and digests associated with this inbox. This cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="ghost" onClick={() => onSetDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={onDelete}
              disabled={deleting}
            >
              {deleting && <Loader2 className="mr-1 h-4 w-4 animate-spin" />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
