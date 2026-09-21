export const MAX_ANALYSIS_RULES = 20;
export const MAX_ANALYSIS_RULE_LENGTH = 300;

export const ANALYSIS_RULE_EXAMPLES = [
  "I want more detail on Finance emails about ISAs, Bitcoin and Ethereum",
  "Keep sports newsletters to a single-sentence summary",
  "Always call out deadlines and dates I need to act on",
];

// Mirrors the backend cap in handler/email.go (maxReanalyseEmails).
export const MAX_REANALYSE_EMAILS = 50;
export const REANALYSE_DAY_OPTIONS = [1, 7, 30];
