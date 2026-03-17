"use client";

import { useCallback, useRef, useState } from "react";

type Permission = "default" | "granted" | "denied" | "unsupported";

function getPermission(): Permission {
  if (typeof window === "undefined" || !("Notification" in window)) {
    return "unsupported";
  }
  return Notification.permission as Permission;
}

export function useEmailNotifications() {
  const [permission, setPermission] = useState<Permission>(getPermission);
  const lastNotifiedId = useRef<string | null>(null);

  const requestPermission = useCallback(async () => {
    if (!("Notification" in window)) {
      setPermission("unsupported");
      return;
    }
    const result = await Notification.requestPermission();
    setPermission(result as Permission);
  }, []);

  const notify = useCallback(
    (emails: { ID: string; FromName: string; FromAddress: string; Subject: string }[]) => {
      if (permission !== "granted") return;
      if (document.hasFocus()) return;
      if (emails.length === 0) return;

      // Avoid duplicate notifications for the same batch
      const latestId = emails[0].ID;
      if (lastNotifiedId.current === latestId) return;
      lastNotifiedId.current = latestId;

      if (emails.length === 1) {
        const e = emails[0];
        new Notification(e.FromName || e.FromAddress, {
          body: e.Subject,
          icon: "/logo.svg",
          tag: `email-${e.ID}`,
        });
      } else {
        new Notification(`${emails.length} new emails`, {
          body: emails.map((e) => e.FromName || e.FromAddress).join(", "),
          icon: "/logo.svg",
          tag: "email-batch",
        });
      }
    },
    [permission]
  );

  return { permission, requestPermission, notify };
}
