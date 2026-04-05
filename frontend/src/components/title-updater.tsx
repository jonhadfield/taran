"use client";

import { useEffect } from "react";
import { apiGet } from "@/lib/api";
import { APP_NAME } from "@/lib/config";
import type { Email, ListResponse } from "@/types/api";

export function TitleUpdater() {
  useEffect(() => {
    async function updateTitle() {
      try {
        const data = await apiGet<ListResponse<Email>>("emails", {
          is_read: "false",
          is_archived: "false",
          limit: "1",
        });
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
