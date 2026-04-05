import { serverFetch } from "@/lib/server-api";
import type { Digest, WeeklySummary, ListResponse } from "@/types/api";
import { GenerateDigestButton } from "./generate-digest-button";
import { GenerateWeeklyButton } from "./generate-weekly-button";
import { DigestsTabs } from "./digests-tabs";
import { DeliveryPrompt } from "./delivery-prompt";

export default async function DigestsPage() {
  let digests: Digest[] = [];
  let digestsTotal = 0;
  let summaries: WeeklySummary[] = [];
  let summariesTotal = 0;

  try {
    const [digestRes, summaryRes] = await Promise.all([
      serverFetch<ListResponse<Digest>>("digests?limit=50"),
      serverFetch<ListResponse<WeeklySummary>>("weekly-summaries?limit=50"),
    ]);
    digests = digestRes.data || [];
    digestsTotal = digestRes.total;
    summaries = summaryRes.data || [];
    summariesTotal = summaryRes.total;
  } catch {
    // Will show empty state
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-balance">Digests</h1>
        <div className="flex items-center gap-2">
          <GenerateWeeklyButton />
          <GenerateDigestButton />
        </div>
      </div>

      {digests.length > 0 && <DeliveryPrompt />}

      <DigestsTabs
        initialDigests={digests}
        initialDigestsTotal={digestsTotal}
        initialSummaries={summaries}
        initialSummariesTotal={summariesTotal}
      />
    </div>
  );
}
