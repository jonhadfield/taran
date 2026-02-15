"use client";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

const filters = ["all", "unread", "starred", "archived"] as const;

interface InboxFiltersProps {
  value: string;
  onChange: (value: string) => void;
}

export function InboxFilters({ value, onChange }: InboxFiltersProps) {
  return (
    <Tabs value={value} onValueChange={onChange}>
      <TabsList>
        {filters.map((f) => (
          <TabsTrigger key={f} value={f} className="capitalize">
            {f}
          </TabsTrigger>
        ))}
      </TabsList>
    </Tabs>
  );
}
