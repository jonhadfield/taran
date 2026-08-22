"use client";

import { useState } from "react";
import { Copy, Check } from "lucide-react";

interface CopyEmailAddressProps {
  emailAddress: string;
}

export function CopyEmailAddress({ emailAddress }: CopyEmailAddressProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(emailAddress);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback for older browsers
      const input = document.createElement("input");
      input.value = emailAddress;
      document.body.appendChild(input);
      input.select();
      document.execCommand("copy");
      document.body.removeChild(input);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <button
      onClick={handleCopy}
      className="inline-flex items-center gap-1.5 rounded-md bg-muted px-3 py-1.5 text-sm font-mono transition-colors hover:bg-muted/80"
    >
      <span className="truncate">{emailAddress}</span>
      {copied ? (
        <Check className="size-3.5 shrink-0 text-success" />
      ) : (
        <Copy className="size-3.5 shrink-0 text-muted-foreground" />
      )}
    </button>
  );
}
