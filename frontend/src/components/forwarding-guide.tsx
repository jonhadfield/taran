"use client";

import { useState } from "react";
import { ChevronDown, ChevronRight, Mail } from "lucide-react";

interface ForwardingGuideProps {
  emailAddress?: string;
}

interface ProviderSection {
  name: string;
  steps: string[];
  tip?: string;
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
    tip: "To forward only newsletters, create a filter: click the search options arrow, enter the sender, click \"Create filter\", then choose \"Forward it to\" your MailBrief address.",
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
    tip: "To forward specific senders only, use Rules: Settings > Mail > Rules > Add new rule. Set the condition to the sender's address and the action to forward.",
  },
  {
    name: "Yahoo Mail",
    steps: [
      "Click the gear icon > More Settings",
      "Select \"Mailboxes\" from the left menu",
      "Click your Yahoo email address",
      "Under \"Forwarding\", enter your MailBrief address",
      "Click Verify and confirm via the verification email",
    ],
    tip: "Yahoo Mail forwards all incoming mail. To be selective, use filters instead: Settings > More Settings > Filters.",
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
    tip: "Create separate rules per newsletter sender for more control over what gets forwarded.",
  },
  {
    name: "ProtonMail",
    steps: [
      "Go to Settings > All settings > Proton Mail > Forwarding",
      "Click \"Add forwarding address\"",
      "Enter your MailBrief address and confirm",
      "Accept the confirmation email sent to your MailBrief inbox",
      "Set up a filter to auto-forward: Filters > Add filter",
    ],
    tip: "ProtonMail requires creating a filter (Sieve) to forward messages. Use conditions like sender address to forward only newsletters.",
  },
  {
    name: "Fastmail",
    steps: [
      "Go to Settings > Rules",
      "Click \"New Rule\"",
      "Set the condition (e.g., from a specific sender)",
      "Set the action to \"Forward to\" your MailBrief address",
      "Save the rule",
    ],
    tip: "Fastmail's rules are powerful — you can match by sender, subject, or header to forward exactly what you want.",
  },
  {
    name: "Zoho Mail",
    steps: [
      "Go to Settings > Mail > Email forwarding",
      "Click \"Add email address\" and enter your MailBrief address",
      "Verify the forwarding address via the confirmation email",
      "Choose whether to keep or delete the original email",
      "Save your settings",
    ],
  },
  {
    name: "Hey.com",
    steps: [
      "Open the email you want to forward",
      "Click the \"...\" menu > Forward",
      "Enter your MailBrief address",
      "Hey doesn't support automatic forwarding — see the tip below",
    ],
    tip: "Hey.com doesn't support auto-forwarding rules. Instead, subscribe to newsletters directly using your MailBrief address.",
  },
  {
    name: "Direct subscription",
    steps: [
      "Visit the newsletter's website",
      "Use your MailBrief address when subscribing",
      "Emails will arrive directly in your MailBrief inbox",
    ],
    tip: "This is often the simplest approach — subscribe directly with your MailBrief address and skip forwarding entirely.",
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
                <div className="px-3 pb-3 space-y-2">
                  <ol className="list-decimal list-inside space-y-1 text-sm text-muted-foreground">
                    {provider.steps.map((step, i) => (
                      <li key={i}>{step}</li>
                    ))}
                  </ol>
                  {provider.tip && (
                    <p className="text-xs text-muted-foreground bg-muted/50 rounded-md px-3 py-2">
                      <span className="font-medium">Tip:</span> {provider.tip}
                    </p>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
