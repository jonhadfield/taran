"use client";

import { useEffect, useRef, useState } from "react";

export interface SSEEvent {
  type: string;
  emailId: string;
}

export type SSEStatus = "connecting" | "connected" | "reconnecting";

export function useEventSource(onEvent: (event: SSEEvent) => void): SSEStatus {
  // Ref pattern: store the latest callback so the SSE effect (which never
  // re-runs) always calls the current version without stale closures.
  const onEventRef = useRef(onEvent);
  const [status, setStatus] = useState<SSEStatus>("connecting");

  // Sync ref on every render (no deps = runs after every render)
  useEffect(() => {
    onEventRef.current = onEvent;
  });

  useEffect(() => {
    let es: EventSource | null = null;
    let retryDelay = 1000;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;
    let unmounted = false;
    let wasConnected = false;

    function connect() {
      if (unmounted) return;

      if (wasConnected) {
        setStatus("reconnecting");
      }

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
        retryDelay = 1000;
        wasConnected = true;
        if (!unmounted) setStatus("connected");
      };

      es.onerror = () => {
        es?.close();
        if (unmounted) return;

        setStatus("reconnecting");
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

  return status;
}
