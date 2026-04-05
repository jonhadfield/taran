"use client";

import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { DigestList } from "./digest-list";
import { WeeklySummaryList } from "./weekly-summary-list";
import type { Digest, WeeklySummary } from "@/types/api";

interface DigestsTabsProps {
  initialDigests: Digest[];
  initialDigestsTotal: number;
  initialSummaries: WeeklySummary[];
  initialSummariesTotal: number;
}

export function DigestsTabs({
  initialDigests,
  initialDigestsTotal,
  initialSummaries,
  initialSummariesTotal,
}: DigestsTabsProps) {
  return (
    <Tabs defaultValue="digests">
      <TabsList>
        <TabsTrigger value="digests">Digests</TabsTrigger>
        <TabsTrigger value="weekly">Weekly Summaries</TabsTrigger>
      </TabsList>
      <TabsContent value="digests">
        <DigestList
          initialDigests={initialDigests}
          initialTotal={initialDigestsTotal}
        />
      </TabsContent>
      <TabsContent value="weekly">
        <WeeklySummaryList
          initialSummaries={initialSummaries}
          initialTotal={initialSummariesTotal}
        />
      </TabsContent>
    </Tabs>
  );
}
