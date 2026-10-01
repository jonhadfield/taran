import * as Sentry from "@sentry/nextjs";
import { sentryIgnoreErrors } from "@/lib/sentry-ignore";

const dsn = process.env.NEXT_PUBLIC_SENTRY_DSN;

Sentry.init({
  dsn,
  enabled: Boolean(dsn),
  environment: process.env.VERCEL_ENV || process.env.NODE_ENV,
  // Keep free-tier usage low; raise if you need more perf signal
  tracesSampleRate: process.env.NODE_ENV === "development" ? 1.0 : 0.05,
  ignoreErrors: sentryIgnoreErrors,
});

export const onRouterTransitionStart = Sentry.captureRouterTransitionStart;
