export const APP_NAME = process.env.NEXT_PUBLIC_APP_NAME || "MailBrief";

/** When false, public pages show auth/reading only — no pitch, process, or demo digest. */
export const SHOW_MARKETING =
  process.env.NEXT_PUBLIC_SHOW_MARKETING !== "false";

