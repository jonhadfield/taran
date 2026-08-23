"use client";

import { useState, useEffect } from "react";
import { apiGet, apiPost } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { Invite, ListResponse } from "@/types/api";
import { Send, CheckCircle2, Clock } from "lucide-react";
import { formatShortDate } from "@/lib/utils";

export function InviteForm() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [invites, setInvites] = useState<Invite[]>([]);

  const fetchInvites = async () => {
    try {
      const res = await apiGet<ListResponse<Invite>>("admin/invites");
      setInvites(res.data ?? []);
    } catch {
      // ignore
    }
  };

  useEffect(() => {
    fetchInvites();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    const trimmed = email.trim().toLowerCase();
    if (!trimmed || !trimmed.includes("@")) {
      setError("Please enter a valid email address");
      return;
    }

    setLoading(true);
    try {
      await apiPost<Invite>("admin/invites", { email: trimmed });
      setSuccess(`Invite sent to ${trimmed}`);
      setEmail("");
      fetchInvites();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to send invite");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Invite User</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="flex gap-3">
            <Input
              type="email"
              placeholder="user@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="flex-1"
            />
            <Button type="submit" disabled={loading}>
              <Send className="mr-2 size-4" />
              {loading ? "Sending..." : "Send Invite"}
            </Button>
          </form>
          {error && (
            <p className="mt-2 text-sm text-destructive">{error}</p>
          )}
          {success && (
            <p className="mt-2 text-sm text-success">{success}</p>
          )}
        </CardContent>
      </Card>

      {invites.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Invites ({invites.length})</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {invites.map((invite) => (
                <div
                  key={invite.ID}
                  className="flex items-center justify-between"
                >
                  <div className="min-w-0">
                    <p className="text-sm font-medium truncate">
                      {invite.Email}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      Invited by {invite.InvitedBy} on{" "}
                      {formatShortDate(invite.CreatedAt)}
                    </p>
                  </div>
                  {invite.AcceptedAt ? (
                    <Badge variant="default" className="gap-1 shrink-0">
                      <CheckCircle2 className="size-3" />
                      Accepted
                    </Badge>
                  ) : (
                    <Badge variant="secondary" className="gap-1 shrink-0">
                      <Clock className="size-3" />
                      Pending
                    </Badge>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
