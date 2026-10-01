/** Known platform noise that should not pollute the Sentry inbox. */
export const sentryIgnoreErrors: Array<string | RegExp> = [
  "Vercel Runtime Timeout Error",
  /Task timed out after \d+ seconds/i,
];
