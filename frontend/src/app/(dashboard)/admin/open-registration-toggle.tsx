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
import { pluralize } from "@/lib/utils";

interface OpenRegistrationState {
  openRegistration: boolean;
  signups: number;
}

export function OpenRegistrationToggle() {
  const [state, setState] = useState<OpenRegistrationState | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    apiGet<OpenRegistrationState>("admin/settings/open-registration")
      .then(setState)
      .catch(() => {});
  }, []);

  const handleToggle = async (checked: boolean) => {
    if (!state) return;
    const previous = state;
    setState({ ...state, openRegistration: checked });
    setSaving(true);
    try {
      const updated = await apiPatch<OpenRegistrationState>(
        "admin/settings/open-registration",
        { OpenRegistration: checked },
      );
      setState(updated);
      toast.success(
        checked
          ? "Open registration is on. Anyone can now sign up."
          : "Open registration is off. The site is invite-only again.",
      );
    } catch (err) {
      setState(previous);
      toast.error(err instanceof Error ? err.message : "Failed to update setting");
    } finally {
      setSaving(false);
    }
  };

  if (!state) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Open registration</CardTitle>
        <CardDescription>
          When on, anyone who signs in with Google or GitHub gets access without
          an invite. People who sign up keep access after you turn it off.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex items-center gap-3">
          <Switch
            id="open-registration"
            checked={state.openRegistration}
            onCheckedChange={handleToggle}
            disabled={saving}
          />
          <Label htmlFor="open-registration">
            {state.openRegistration ? "Open to everyone" : "Invite-only"}
          </Label>
        </div>
        <p className="text-sm text-muted-foreground">
          <span className="font-medium text-foreground">{state.signups}</span>{" "}
          {pluralize(state.signups, "person", "people")} signed up through open
          registration
        </p>
      </CardContent>
    </Card>
  );
}
