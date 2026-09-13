import { NextRequest, NextResponse } from "next/server";
import { getSessionCookie } from "better-auth/cookies";

export async function proxy(request: NextRequest) {
  const sessionCookie = getSessionCookie(request);
  const { pathname } = request.nextUrl;

  // Don't redirect from /login to / based on cookie alone — the cookie may be
  // expired or rotated. Let the login page check session validity client-side.

  if (sessionCookie && pathname === "/not-invited") {
    // Allow authenticated users to see the not-invited page
    return NextResponse.next();
  }

  if (!sessionCookie && !pathname.startsWith("/login") && !pathname.startsWith("/not-invited") && !pathname.startsWith("/api/auth") && !pathname.startsWith("/shared")) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}

// Skip Next.js internals, API routes, and the icon files the login page loads —
// otherwise signed-out visitors are redirected and receive the login page's
// HTML in place of the icon. File names are matched exactly.
export const config = {
  matcher: [
    "/((?!_next/static|_next/image|api/|(?:favicon\\.ico|icon\\.svg|apple-icon\\.png|logo\\.svg|logo-192\\.png|logo-512\\.png)$).*)",
  ],
};
