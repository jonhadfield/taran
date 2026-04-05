"use client";

import { useState, useEffect } from "react";
import { apiGet, apiPatch } from "@/lib/api";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { toast } from "sonner";

export function WaitlistToggle() {
  const [enabled, setEnabled] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiGet<{ waitlistEnabled: boolean }>("admin/settings/waitlist")
      .then((res) => setEnabled(res.waitlistEnabled))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const handleToggle = async (checked: boolean) => {
    setEnabled(checked);
    try {
      await apiPatch("admin/settings/waitlist", {
        WaitlistEnabled: checked,
      });
      toast.success(
        checked
          ? "Waitlist is now open — users can request access"
          : "Waitlist is now closed"
      );
    } catch (err) {
      setEnabled(!checked);
      toast.error(
        err instanceof Error ? err.message : "Failed to update setting"
      );
    }
  };

  if (loading) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Waitlist</CardTitle>
        <CardDescription>
          When enabled, non-invited users can request access from the login
          screen.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-3">
          <Switch
            id="waitlist-enabled"
            checked={enabled}
            onCheckedChange={handleToggle}
          />
          <Label htmlFor="waitlist-enabled">
            {enabled ? "Waitlist open" : "Waitlist closed"}
          </Label>
        </div>
      </CardContent>
    </Card>
  );
}
