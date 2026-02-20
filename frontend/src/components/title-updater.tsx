"use client";

import { useEffect } from "react";
import { APP_NAME } from "@/lib/config";

export function TitleUpdater() {
  useEffect(() => {
    async function updateTitle() {
      try {
        const res = await fetch("/api/proxy/emails?is_read=false&is_archived=false&limit=1");
        if (!res.ok) return;
        const data = await res.json();
        const count = data.total ?? 0;
        document.title = count > 0 ? `(${count}) ${APP_NAME}` : APP_NAME;
      } catch {
        // Silently ignore — title stays as-is
      }
    }

    updateTitle();
    const timer = setInterval(updateTitle, 60_000);

    return () => clearInterval(timer);
  }, []);

  return null;
}
