import { NextRequest, NextResponse } from "next/server";
import { getSessionCookie } from "better-auth/cookies";

export async function proxy(request: NextRequest) {
  const sessionCookie = getSessionCookie(request);
  const { pathname } = request.nextUrl;

  if (sessionCookie && pathname === "/login") {
    return NextResponse.redirect(new URL("/", request.url));
  }

  if (sessionCookie && pathname === "/not-invited") {
    // Allow authenticated users to see the not-invited page
    return NextResponse.next();
  }

  if (!sessionCookie && !pathname.startsWith("/login") && !pathname.startsWith("/not-invited") && !pathname.startsWith("/api/auth") && !pathname.startsWith("/shared")) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|logo.svg|api/).*)"],
};
