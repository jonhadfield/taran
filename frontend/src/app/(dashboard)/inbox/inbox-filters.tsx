"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

const filters = ["all", "unread", "starred", "archived"] as const;

export function InboxFilters() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const current = searchParams.get("filter") || "all";

  const onChange = (value: string) => {
    const params = new URLSearchParams(searchParams);
    if (value === "all") {
      params.delete("filter");
    } else {
      params.set("filter", value);
    }
    const qs = params.toString();
    router.replace(`/inbox${qs ? `?${qs}` : ""}`);
  };

  return (
    <Tabs value={current} onValueChange={onChange}>
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
