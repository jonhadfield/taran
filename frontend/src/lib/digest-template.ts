export type DigestSectionId = "highlights" | "topics" | "emails";
export type DigestDensity = "comfortable" | "compact";

export const DEFAULT_DIGEST_SECTION_ORDER: DigestSectionId[] = [
  "highlights",
  "topics",
  "emails",
];

const VALID = new Set<DigestSectionId>(["highlights", "topics", "emails"]);

export function normalizeDigestSectionOrder(
  order: string[] | null | undefined,
): DigestSectionId[] {
  const seen = new Set<DigestSectionId>();
  const out: DigestSectionId[] = [];
  for (const raw of order || []) {
    if (!VALID.has(raw as DigestSectionId) || seen.has(raw as DigestSectionId)) {
      continue;
    }
    const id = raw as DigestSectionId;
    seen.add(id);
    out.push(id);
  }
  for (const id of DEFAULT_DIGEST_SECTION_ORDER) {
    if (!seen.has(id)) out.push(id);
  }
  return out;
}

export function normalizeDigestDensity(
  density: string | null | undefined,
): DigestDensity {
  return density === "compact" ? "compact" : "comfortable";
}

export function moveDigestSection(
  order: DigestSectionId[],
  index: number,
  direction: -1 | 1,
): DigestSectionId[] {
  const next = [...order];
  const target = index + direction;
  if (target < 0 || target >= next.length) return next;
  [next[index], next[target]] = [next[target], next[index]];
  return next;
}
