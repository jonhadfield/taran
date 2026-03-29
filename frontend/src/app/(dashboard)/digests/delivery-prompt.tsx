"use client";

import { useEffect, useState } from "react";
import { apiGet } from "@/lib/api";
import Link from "next/link";
import { Mail } from "lucide-react";
import type { UserPreference } from "@/types/api";

export function DeliveryPrompt() {
  const [show, setShow] = useState(false);

  useEffect(() => {
    apiGet<UserPreference>("preferences")
      .then((pref) => {
        if (!pref.DigestEmail) setShow(true);
      })
      .catch(() => {});
  }, []);

  if (!show) return null;

  return (
    <div className="flex items-center gap-3 rounded-lg border border-dashed p-3 text-sm text-muted-foreground">
      <Mail className="h-4 w-4 shrink-0" />
      <span>
        Want digests delivered to your inbox?{" "}
        <Link href="/settings" className="underline text-foreground">
          Enable in Settings
        </Link>
      </span>
    </div>
  );
}
