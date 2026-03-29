import { serverFetch } from "@/lib/server-api";
import type { Digest, ListResponse } from "@/types/api";
import { GenerateDigestButton } from "./generate-digest-button";
import { DigestList } from "./digest-list";
import { DeliveryPrompt } from "./delivery-prompt";

export default async function DigestsPage() {
  let digests: Digest[] = [];
  let total = 0;

  try {
    const res = await serverFetch<ListResponse<Digest>>("digests?limit=50");
    digests = res.data || [];
    total = res.total;
  } catch {
    // Will show empty state
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-balance">Digests</h1>
        <GenerateDigestButton />
      </div>

      {digests.length > 0 && <DeliveryPrompt />}

      <DigestList initialDigests={digests} initialTotal={total} />
    </div>
  );
}
