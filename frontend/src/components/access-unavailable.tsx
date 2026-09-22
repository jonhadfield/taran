import Image from "next/image";
import Link from "next/link";
import { APP_NAME } from "@/lib/config";

/**
 * Shown when the access check itself fails, so we don't know whether the user
 * may use the app. Deliberately not the "not invited" page: the usual cause is
 * a brief backend problem, not a missing invite.
 */
export function AccessUnavailable() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 px-6 text-center">
      <Image src="/icon.svg" alt="" width={48} height={48} priority />
      <h1 className="text-xl font-semibold">{APP_NAME} is temporarily unavailable</h1>
      <p className="max-w-sm text-sm text-muted-foreground">
        We couldn&apos;t reach the service just now. Please try again in a moment.
      </p>
      <Link
        href="/"
        className="rounded-md border px-4 py-2 text-sm font-medium hover:bg-muted"
      >
        Try again
      </Link>
    </main>
  );
}
