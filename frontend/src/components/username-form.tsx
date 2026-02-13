"use client";

import { useCallback, useEffect, useState } from "react";
import { apiGet, apiPost, ApiError } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Check, Loader2, X } from "lucide-react";
import type { EmailAccount } from "@/types/api";

const EMAIL_DOMAIN = process.env.NEXT_PUBLIC_EMAIL_DOMAIN || "mailbrief.io";

interface UsernameFormProps {
  onSuccess: () => void;
}

export function UsernameForm({ onSuccess }: UsernameFormProps) {
  const [username, setUsername] = useState("");
  const [checking, setChecking] = useState(false);
  const [available, setAvailable] = useState<boolean | null>(null);
  const [reason, setReason] = useState("");
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");

  const checkAvailability = useCallback(async (value: string) => {
    if (value.length < 3) {
      setAvailable(null);
      setReason("");
      return;
    }

    setChecking(true);
    try {
      const res = await apiGet<{ available: boolean; reason?: string }>(
        "accounts/check-username",
        { username: value }
      );
      setAvailable(res.available);
      setReason(res.reason || "");
    } catch {
      setAvailable(null);
      setReason("");
    } finally {
      setChecking(false);
    }
  }, []);

  useEffect(() => {
    const value = username.toLowerCase().trim();
    if (!value) {
      setAvailable(null);
      setReason("");
      return;
    }

    const timer = setTimeout(() => checkAvailability(value), 400);
    return () => clearTimeout(timer);
  }, [username, checkAvailability]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!available) return;

    setCreating(true);
    setError("");

    try {
      await apiPost<EmailAccount>("accounts", {
        Username: username.toLowerCase().trim(),
      });
      onSuccess();
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setAvailable(false);
        setReason("");
      } else {
        setError(err instanceof Error ? err.message : "Failed to create inbox");
      }
    } finally {
      setCreating(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="username">Username</Label>
        <div className="flex items-center gap-2">
          <div className="relative flex-1">
            <Input
              id="username"
              placeholder="your-username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="pr-8"
              autoComplete="off"
              autoFocus
            />
            {username.length >= 3 && (
              <div className="absolute right-2 top-1/2 -translate-y-1/2">
                {checking ? (
                  <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                ) : available === true ? (
                  <Check className="h-4 w-4 text-green-500" />
                ) : available === false ? (
                  <X className="h-4 w-4 text-red-500" />
                ) : null}
              </div>
            )}
          </div>
          <span className="text-sm text-muted-foreground whitespace-nowrap">
            @{EMAIL_DOMAIN}
          </span>
        </div>
        {available === false && (
          <p className="text-sm text-destructive">
            {reason || "This username is already taken"}
          </p>
        )}
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <Button type="submit" disabled={!available || creating}>
        {creating && <Loader2 className="mr-1 h-4 w-4 animate-spin" />}
        Create Inbox
      </Button>
    </form>
  );
}
