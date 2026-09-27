/**
 * Environment-aware logger.
 *
 * - Development: full console output
 * - Production: errors only, forwarded to Sentry when a DSN is set
 */

import * as Sentry from "@sentry/nextjs";

type LogLevel = "debug" | "info" | "warn" | "error";

interface LogContext {
  [key: string]: unknown;
}

class Logger {
  private isDevelopment = process.env.NODE_ENV === "development";
  private isTest = process.env.NODE_ENV === "test";

  debug(message: string, context?: LogContext): void {
    if (this.isDevelopment) {
      console.log(`[DEBUG] ${message}`, context || "");
    }
  }

  info(message: string, context?: LogContext): void {
    if (this.isDevelopment) {
      console.log(`[INFO] ${message}`, context || "");
    }
  }

  warn(message: string, context?: LogContext): void {
    if (this.isDevelopment) {
      console.warn(`[WARN] ${message}`, context || "");
    }
  }

  /**
   * Always logged. In production, also sent to Sentry when configured.
   */
  error(message: string, error?: Error | unknown, context?: LogContext): void {
    if (this.isDevelopment || this.isTest) {
      console.error(`[ERROR] ${message}`, error || "", context || "");
    } else {
      // Avoid dumping potentially sensitive context into production logs
      console.error(`[ERROR] ${message}`);
    }

    if (!this.isTest && process.env.NEXT_PUBLIC_SENTRY_DSN) {
      Sentry.withScope((scope) => {
        scope.setLevel("error");
        scope.setTag("source", "logger");
        scope.setExtra("message", message);
        if (context) {
          scope.setExtra("context", context);
        }
        if (error instanceof Error) {
          Sentry.captureException(error);
        } else if (error !== undefined) {
          Sentry.captureException(new Error(message), {
            extra: { cause: error },
          });
        } else {
          Sentry.captureMessage(message, "error");
        }
      });
    }
  }

  failed(operation: string, error: Error | unknown, context?: LogContext): void {
    this.error(`Failed to ${operation}`, error, context);
  }

  success(operation: string, context?: LogContext): void {
    if (this.isDevelopment) {
      console.log(`[SUCCESS] ${operation}`, context || "");
    }
  }
}

export const logger = new Logger();

export type { LogLevel, LogContext };
