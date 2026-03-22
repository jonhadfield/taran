"use client";

import { useState } from "react";
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
}

export function OnboardingChecklist({
  emailAddress,
  hasEmails,
  hasDigest,
}: OnboardingChecklistProps) {
  const [dismissed, setDismissed] = useState(() => {
    if (typeof window === "undefined") return false;
    return localStorage.getItem("onboarding-checklist-dismissed") === "1";
  });

  if (dismissed) return null;

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
      done: false, // Always shown as a suggestion until dismissed
      href: "/settings",
      icon: <Settings className="size-4" />,
    },
  ];

  const completedCount = items.filter((i) => i.done).length;
  const allDone = completedCount >= 3; // 3 of 4 (settings is always "todo" until dismissed)

  const handleDismiss = () => {
    localStorage.setItem("onboarding-checklist-dismissed", "1");
    setDismissed(true);
  };

  return (
    <Card className="border-primary/20 bg-primary/[0.02]">
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
                <Check className="size-4 text-green-500" />
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

        {!allDone && emailAddress && (
          <div className="pt-1">
            <p className="text-xs text-muted-foreground">Your inbox address:</p>
            <div className="mt-1">
              <CopyEmailAddress emailAddress={emailAddress} />
            </div>
          </div>
        )}

        {allDone && (
          <div className="pt-1 text-center">
            <p className="text-sm text-muted-foreground">
              You&apos;re all set! You can dismiss this checklist.
            </p>
            <Button variant="outline" size="sm" className="mt-2" onClick={handleDismiss}>
              Dismiss
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
