import { NextRequest, NextResponse } from "next/server";
import { getSessionCookie } from "better-auth/cookies";
import { loginPathWithCallback } from "@/lib/auth-callback";

export async function proxy(request: NextRequest) {
  const sessionCookie = getSessionCookie(request);
  const { pathname, search } = request.nextUrl;

  // Don't redirect from /login to / based on cookie alone — the cookie may be
  // expired or rotated. Let the login page check session validity client-side.

  if (sessionCookie && pathname === "/not-invited") {
    // Allow authenticated users to see the not-invited page
    return NextResponse.next();
  }

  if (
    !sessionCookie &&
    !pathname.startsWith("/login") &&
    !pathname.startsWith("/not-invited") &&
    !pathname.startsWith("/api/auth") &&
    !pathname.startsWith("/shared")
  ) {
    // Preserve the deep link (e.g. /inbox/{id} from a digest "View full email"
    // link) so login can send the user back after OAuth.
    const loginPath = loginPathWithCallback(pathname + search);
    return NextResponse.redirect(new URL(loginPath, request.url));
  }

  // Expose the path to server layouts so mid-session expiry can redirect with
  // the same callbackURL behaviour.
  const requestHeaders = new Headers(request.headers);
  requestHeaders.set("x-pathname", pathname + search);
  return NextResponse.next({
    request: { headers: requestHeaders },
  });
}

// Skip Next.js internals, API routes, the Sentry tunnel, and the public assets
// the login page loads — otherwise signed-out visitors are redirected and
// receive the login page's HTML in place of the asset. File names are matched
// exactly.
export const config = {
  matcher: [
    "/((?!_next/static|_next/image|api/|sentry-tunnel|(?:favicon\\.ico|icon\\.svg|apple-icon\\.png|logo\\.svg|logo-192\\.png|logo-512\\.png|digest-flow\\.png)$).*)",
  ],
};
