"use client";

import { useEffect, useState } from "react";
import { apiGet, apiDelete, apiPatch } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Loader2, Trash2 } from "lucide-react";
import type { EmailAccount, ListResponse, UserPreference } from "@/types/api";
import { CopyButton } from "./copy-button";
import { SignOutButton } from "./sign-out-button";
import { UsernameForm } from "@/components/username-form";

export default function SettingsPage() {
  const [accounts, setAccounts] = useState<EmailAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Preferences
  const [digestEmail, setDigestEmail] = useState(false);
  const [prefLoading, setPrefLoading] = useState(true);
  const [prefSaving, setPrefSaving] = useState(false);

  // Delete confirmation
  const [deleteTarget, setDeleteTarget] = useState<EmailAccount | null>(null);
  const [deleting, setDeleting] = useState(false);

  const fetchAccounts = async () => {
    try {
      const res = await apiGet<ListResponse<EmailAccount>>("accounts");
      setAccounts(res.data || []);
      setError("");
    } catch {
      setError("Failed to load accounts");
    } finally {
      setLoading(false);
    }
  };

  const fetchPreferences = async () => {
    try {
      const pref = await apiGet<UserPreference>("preferences");
      setDigestEmail(pref.DigestEmail);
    } catch {
      // Defaults to false
    } finally {
      setPrefLoading(false);
    }
  };

  useEffect(() => {
    fetchAccounts();
    fetchPreferences();
  }, []);

  const handleToggleDigestEmail = async (checked: boolean) => {
    setPrefSaving(true);
    setDigestEmail(checked);
    try {
      const updated = await apiPatch<UserPreference>("preferences", {
        DigestEmail: checked,
      });
      setDigestEmail(updated.DigestEmail);
    } catch {
      setDigestEmail(!checked);
    } finally {
      setPrefSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);

    try {
      await apiDelete(`accounts/${deleteTarget.ID}`);
      setDeleteTarget(null);
      await fetchAccounts();
    } catch {
      setDeleteTarget(null);
    } finally {
      setDeleting(false);
    }
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <h1 className="text-2xl font-bold">Settings</h1>

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
              <UsernameForm onSuccess={fetchAccounts} />
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
                  onClick={() => setDeleteTarget(account)}
                >
                  <Trash2 className="h-4 w-4 text-muted-foreground" />
                </Button>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Digest Delivery</CardTitle>
          <CardDescription>
            Get your daily digest delivered to your email inbox
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <Label htmlFor="digest-email" className="flex flex-col items-start gap-1">
              <span>Email delivery</span>
              <span className="text-sm font-normal text-muted-foreground">
                Receive your digest as an email each day
              </span>
            </Label>
            <Switch
              id="digest-email"
              checked={digestEmail}
              onCheckedChange={handleToggleDigestEmail}
              disabled={prefLoading || prefSaving}
            />
          </div>
        </CardContent>
      </Card>

      <Separator />

      <Card>
        <CardHeader>
          <CardTitle>Account</CardTitle>
          <CardDescription>Manage your account settings</CardDescription>
        </CardHeader>
        <CardContent>
          <SignOutButton />
        </CardContent>
      </Card>

      <Dialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Inbox</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete{" "}
              <span className="font-medium">
                {deleteTarget?.EmailAddress}
              </span>
              ? This inbox will stop receiving emails.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleting}
            >
              {deleting && <Loader2 className="mr-1 h-4 w-4 animate-spin" />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
