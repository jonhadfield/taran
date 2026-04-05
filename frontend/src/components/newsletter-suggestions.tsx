"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ExternalLink } from "lucide-react";

const NEWSLETTER_SUGGESTIONS = [
  { name: "Morning Brew", description: "Daily business news", url: "https://www.morningbrew.com/daily" },
  { name: "TLDR", description: "Byte-sized tech news", url: "https://tldr.tech/" },
  { name: "The Hustle", description: "Business & tech trends", url: "https://thehustle.co/" },
  { name: "Superhuman", description: "AI news & insights", url: "https://www.superhuman.ai/" },
];

function SuggestionList() {
  return (
    <div className="grid gap-2">
      {NEWSLETTER_SUGGESTIONS.map((nl) => (
        <a
          key={nl.name}
          href={nl.url}
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center justify-between rounded-lg border p-3 hover:bg-muted/50 transition-colors"
        >
          <div>
            <p className="text-sm font-medium">{nl.name}</p>
            <p className="text-xs text-muted-foreground">{nl.description}</p>
          </div>
          <ExternalLink className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        </a>
      ))}
    </div>
  );
}

/** Render inside a Card (settings page) or inline (onboarding). */
export function NewsletterSuggestions({
  emailAddress,
  inline,
}: {
  emailAddress?: string;
  inline?: boolean;
}) {
  if (inline) {
    return (
      <div className="space-y-3">
        <p className="text-sm font-medium">Popular newsletters to try:</p>
        <SuggestionList />
      </div>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Discover Newsletters</CardTitle>
        <CardDescription>
          Popular newsletters to subscribe to your inbox
          {emailAddress && (
            <> (<code className="text-xs">{emailAddress}</code>)</>
          )}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <SuggestionList />
      </CardContent>
    </Card>
  );
}
