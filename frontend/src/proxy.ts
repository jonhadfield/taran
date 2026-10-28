import { NextRequest, NextResponse } from "next/server";
import { getSessionCookie } from "better-auth/cookies";
import { loginPathWithCallback } from "@/lib/auth-callback";
import { SHOW_MARKETING } from "@/lib/config";

export function isPublicPath(pathname: string): boolean {
  if (pathname.startsWith("/login")) return true;
  if (pathname.startsWith("/not-invited")) return true;
  if (pathname.startsWith("/api/auth")) return true;
  if (pathname.startsWith("/shared")) return true;
  // `/` is rewritten to `/home` for signed-out visitors. `/home` stays public
  // so the rewrite target is reachable; robots.txt keeps it out of the index.
  if (SHOW_MARKETING && (pathname === "/" || pathname === "/home")) return true;
  return false;
}

export async function proxy(request: NextRequest) {
  const sessionCookie = getSessionCookie(request);
  const { pathname, search } = request.nextUrl;

  // Don't redirect from /login to / based on cookie alone — the cookie may be
  // expired or rotated. Let the login page check session validity client-side.

  if (sessionCookie && pathname === "/not-invited") {
    // Allow authenticated users to see the not-invited page
    return NextResponse.next();
  }

  // Internal marketing route. Signed-out / is rewritten here; send anyone
  // who requests /home directly back to the canonical homepage.
  if (pathname === "/home" && (sessionCookie || SHOW_MARKETING)) {
    return NextResponse.redirect(
      new URL("/", request.url),
      sessionCookie ? 307 : 308,
    );
  }

  if (!sessionCookie && !isPublicPath(pathname)) {
    // Preserve the deep link (e.g. /inbox/{id} from a digest "View full email"
    // link) so login can send the user back after OAuth.
    const loginPath = loginPathWithCallback(pathname + search);
    return NextResponse.redirect(new URL(loginPath, request.url));
  }

  if (!sessionCookie && SHOW_MARKETING && pathname === "/") {
    return NextResponse.rewrite(new URL("/home", request.url));
  }

  // Expose the path to server layouts so mid-session expiry can redirect with
  // the same callbackURL behaviour.
  const requestHeaders = new Headers(request.headers);
  requestHeaders.set("x-pathname", pathname + search);
  return NextResponse.next({
    request: { headers: requestHeaders },
  });
}

// Skip Next.js internals, API routes, the Sentry tunnel, crawler files, and the
// public assets the login page loads — otherwise signed-out visitors are
// redirected and receive the login page's HTML in place of the asset. File
// names are matched exactly.
export const config = {
  matcher: [
    "/((?!_next/static|_next/image|api/|sentry-tunnel|\\.well-known/|(?:favicon\\.ico|icon\\.svg|apple-icon\\.png|logo\\.svg|logo-192\\.png|logo-512\\.png|digest-flow\\.png|robots\\.txt|sitemap\\.xml)$).*)",
  ],
};
