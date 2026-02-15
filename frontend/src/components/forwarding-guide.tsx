"use client";

import { useState } from "react";
import { ChevronDown, ChevronRight, Mail } from "lucide-react";

interface ForwardingGuideProps {
  emailAddress?: string;
}

interface ProviderSection {
  name: string;
  steps: string[];
}

const PROVIDERS: ProviderSection[] = [
  {
    name: "Gmail",
    steps: [
      "Open Gmail Settings (gear icon) > See all settings",
      'Go to the "Forwarding and POP/IMAP" tab',
      'Click "Add a forwarding address"',
      "Enter your MailBrief address and click Next",
      "Confirm the forwarding address when prompted",
    ],
  },
  {
    name: "Outlook / Hotmail",
    steps: [
      "Go to Settings > Mail > Forwarding",
      "Check \"Enable forwarding\"",
      "Enter your MailBrief address",
      "Optionally keep a copy of forwarded messages",
      "Click Save",
    ],
  },
  {
    name: "Apple Mail (iCloud)",
    steps: [
      "Go to iCloud.com > Mail > Settings (gear icon)",
      "Click Rules > Add a Rule",
      'Set condition: "If a message is from" any sender',
      'Set action: "Forward to" your MailBrief address',
      "Click Done",
    ],
  },
  {
    name: "Direct subscription",
    steps: [
      "Visit the newsletter's website",
      "Use your MailBrief address when subscribing",
      "Emails will arrive directly in your MailBrief inbox",
    ],
  },
];

export function ForwardingGuide({ emailAddress }: ForwardingGuideProps) {
  const [openSection, setOpenSection] = useState<string | null>(null);

  const toggle = (name: string) => {
    setOpenSection((prev) => (prev === name ? null : name));
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 text-sm font-medium">
        <Mail className="size-4 text-muted-foreground" />
        How to forward your newsletters
      </div>
      {emailAddress && (
        <p className="text-xs text-muted-foreground">
          Forward to: <span className="font-mono font-medium">{emailAddress}</span>
        </p>
      )}
      <div className="rounded-lg border divide-y">
        {PROVIDERS.map((provider) => {
          const isOpen = openSection === provider.name;
          return (
            <div key={provider.name}>
              <button
                onClick={() => toggle(provider.name)}
                className="flex w-full items-center justify-between p-3 text-sm font-medium hover:bg-accent/50 transition-colors"
              >
                {provider.name}
                {isOpen ? (
                  <ChevronDown className="size-4 text-muted-foreground" />
                ) : (
                  <ChevronRight className="size-4 text-muted-foreground" />
                )}
              </button>
              {isOpen && (
                <div className="px-3 pb-3">
                  <ol className="list-decimal list-inside space-y-1 text-sm text-muted-foreground">
                    {provider.steps.map((step, i) => (
                      <li key={i}>{step}</li>
                    ))}
                  </ol>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
