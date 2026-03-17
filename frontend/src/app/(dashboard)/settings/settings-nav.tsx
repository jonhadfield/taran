"use client";

import { useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";

export interface SettingsSection {
  id: string;
  label: string;
  group?: string;
}

interface SettingsNavProps {
  sections: SettingsSection[];
}

export function SettingsNav({ sections }: SettingsNavProps) {
  const [activeId, setActiveId] = useState(sections[0]?.id ?? "");
  const isClicking = useRef(false);
  const pillRef = useRef<HTMLButtonElement | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const els = sections
      .map((s) => document.getElementById(s.id))
      .filter(Boolean) as HTMLElement[];

    if (els.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (isClicking.current) return;
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id);
            break;
          }
        }
      },
      { rootMargin: "-10% 0px -80% 0px", threshold: 0 }
    );

    els.forEach((el) => observer.observe(el));
    return () => observer.disconnect();
  }, [sections]);

  useEffect(() => {
    if (pillRef.current && scrollRef.current) {
      const container = scrollRef.current;
      const pill = pillRef.current;
      const left = pill.offsetLeft - container.offsetLeft - 16;
      container.scrollTo({ left, behavior: "smooth" });
    }
  }, [activeId]);

  const scrollTo = (id: string) => {
    const el = document.getElementById(id);
    if (!el) return;

    isClicking.current = true;
    setActiveId(id);

    el.scrollIntoView({ behavior: "smooth", block: "start" });

    setTimeout(() => {
      isClicking.current = false;
    }, 800);
  };

  return (
    <div className="sticky top-14 z-10 -mx-4 lg:-mx-6 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div
        ref={scrollRef}
        className="flex gap-1 overflow-x-auto px-4 lg:px-6 py-2 scrollbar-hide"
      >
        {sections.map((s, i) => {
          const prevGroup = i > 0 ? sections[i - 1].group : undefined;
          const showSep = s.group && s.group !== prevGroup;
          return (
            <span key={s.id} className="contents">
              {showSep && i > 0 && (
                <span className="shrink-0 w-px h-5 bg-border self-center mx-1" />
              )}
              <button
                ref={s.id === activeId ? pillRef : null}
                onClick={() => scrollTo(s.id)}
                className={cn(
                  "shrink-0 rounded-full px-3 py-1.5 text-xs font-medium transition-colors",
                  s.id === activeId
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-accent hover:text-foreground"
                )}
              >
                {s.label}
              </button>
            </span>
          );
        })}
      </div>
    </div>
  );
}
