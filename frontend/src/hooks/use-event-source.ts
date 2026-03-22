"use client";

import { useEffect, useRef } from "react";

export interface SSEEvent {
  type: string;
  emailId: string;
}

export function useEventSource(onEvent: (event: SSEEvent) => void): void {
  const onEventRef = useRef(onEvent);

  useEffect(() => {
    onEventRef.current = onEvent;
  });

  useEffect(() => {
    let es: EventSource | null = null;
    let retryDelay = 1000;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;
    let unmounted = false;

    function connect() {
      if (unmounted) return;

      es = new EventSource("/api/events");

      es.onmessage = (e) => {
        try {
          const data = JSON.parse(e.data) as SSEEvent;
          onEventRef.current(data);
        } catch {
          // Ignore unparseable events (e.g. keepalive comments)
        }
      };

      es.onopen = () => {
        retryDelay = 1000; // Reset backoff on successful connection
      };

      es.onerror = () => {
        es?.close();
        if (unmounted) return;

        // Reconnect with exponential backoff (max 30s)
        retryTimer = setTimeout(connect, retryDelay);
        retryDelay = Math.min(retryDelay * 2, 30000);
      };
    }

    connect();

    return () => {
      unmounted = true;
      es?.close();
      if (retryTimer) clearTimeout(retryTimer);
    };
  }, []);
}
