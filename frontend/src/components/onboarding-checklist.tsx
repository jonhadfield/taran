"use client";

import { useState, useSyncExternalStore } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Check, Circle, Mail, BookOpen, Settings, X, ExternalLink } from "lucide-react";
import Link from "next/link";
import { CopyEmailAddress } from "@/components/copy-email-address";

interface ChecklistItem {
  id: string;
  label: string;
  description: string;
  done: boolean;
  href?: string;
  icon: React.ReactNode;
}

interface OnboardingChecklistProps {
  emailAddress: string;
  hasEmails: boolean;
  hasDigest: boolean;
  /** null while the preferences lookup is still in flight. */
  hasConfiguredPreferences: boolean | null;
}

export function OnboardingChecklist({
  emailAddress,
  hasEmails,
  hasDigest,
  hasConfiguredPreferences,
}: OnboardingChecklistProps) {
  const wasDismissed = useSyncExternalStore(
    (cb) => { window.addEventListener("storage", cb); return () => window.removeEventListener("storage", cb); },
    () => localStorage.getItem("onboarding-checklist-dismissed") === "1",
    () => false,
  );
  const [dismissed, setDismissed] = useState(false);

  const items: ChecklistItem[] = [
    {
      id: "inbox",
      label: "Create your inbox",
      description: "Choose a username for your email address.",
      done: !!emailAddress,
      icon: <Mail className="size-4" />,
    },
    {
      id: "email",
      label: "Receive your first email",
      description: emailAddress
        ? `Forward a newsletter to ${emailAddress}`
        : "Forward a newsletter to your inbox address.",
      done: hasEmails,
      href: "/inbox",
      icon: <Mail className="size-4" />,
    },
    {
      id: "digest",
      label: "Generate your first digest",
      description: "Go to Digests and generate a summary of your emails.",
      done: hasDigest,
      href: "/digests",
      icon: <BookOpen className="size-4" />,
    },
    {
      id: "settings",
      label: "Configure your preferences",
      description: "Set your digest schedule, timezone, and notification preferences.",
      done: hasConfiguredPreferences === true,
      href: "/settings/digest",
      icon: <Settings className="size-4" />,
    },
  ];

  const completedCount = items.filter((i) => i.done).length;
  const allDone = completedCount === items.length;

  // Render nothing until the preferences lookup resolves. It decides the last
  // item, so rendering early shows a stale "3 of 4" to people who finished
  // onboarding long ago, then rips it away once the answer arrives.
  if (hasConfiguredPreferences === null) return null;

  // Onboarding is finished, so there is nothing to guide. This also means an
  // established account never depends on the dismissal flag below, which lives
  // in localStorage and is therefore lost on a new browser or cleared site data.
  if (allDone) return null;

  if (wasDismissed || dismissed) return null;

  const handleDismiss = () => {
    localStorage.setItem("onboarding-checklist-dismissed", "1");
    setDismissed(true);
  };

  // Keep the card surface and let the border carry the accent. A 2% primary
  // wash replaced bg-card entirely, which left this panel sitting at the page
  // colour while every other card steps up from it.
  return (
    <Card className="border-primary/20">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base">
            Getting started
            <span className="ml-2 text-sm font-normal text-muted-foreground">
              {completedCount} of {items.length} complete
            </span>
          </CardTitle>
          <Button
            variant="ghost"
            size="sm"
            className="h-7 w-7 p-0 text-muted-foreground hover:text-foreground"
            onClick={handleDismiss}
            aria-label="Dismiss checklist"
          >
            <X className="size-4" />
          </Button>
        </div>
        {/* Progress bar */}
        <div className="h-1.5 w-full rounded-full bg-muted mt-2">
          <div
            className="h-1.5 rounded-full bg-primary transition-all duration-500"
            style={{ width: `${(completedCount / items.length) * 100}%` }}
          />
        </div>
      </CardHeader>
      <CardContent className="space-y-2 pt-0">
        {items.map((item) => (
          <div
            key={item.id}
            className={`flex items-start gap-3 rounded-lg p-2.5 transition-colors ${
              item.done ? "opacity-60" : "hover:bg-accent/50"
            }`}
          >
            <div className="mt-0.5">
              {item.done ? (
                <Check className="size-4 text-success" />
              ) : (
                <Circle className="size-4 text-muted-foreground" />
              )}
            </div>
            <div className="flex-1 min-w-0">
              <p className={`text-sm font-medium ${item.done ? "line-through" : ""}`}>
                {item.label}
              </p>
              <p className="text-xs text-muted-foreground">{item.description}</p>
            </div>
            {!item.done && item.href && (
              <Link
                href={item.href}
                className="shrink-0 inline-flex items-center gap-1 text-xs text-primary hover:underline mt-0.5"
              >
                Go
                <ExternalLink className="size-3" />
              </Link>
            )}
          </div>
        ))}

        {emailAddress && (
          <div className="pt-1">
            <p className="text-xs text-muted-foreground">Your inbox address:</p>
            <div className="mt-1">
              <CopyEmailAddress emailAddress={emailAddress} />
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
