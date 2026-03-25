import { auth } from "@/lib/auth";
import { toNextJsHandler } from "better-auth/next-js";
import { NextRequest, NextResponse } from "next/server";

const { GET: authGET, POST } = toNextJsHandler(auth);

// Wrap GET to fix cache headers on session endpoint
// Better Auth returns Cache-Control: public which can leak session data in shared caches
async function GET(req: NextRequest) {
  const res = await authGET(req);
  if (res instanceof NextResponse || res instanceof Response) {
    const newRes = new NextResponse(res.body, {
      status: res.status,
      headers: res.headers,
    });
    newRes.headers.set("Cache-Control", "no-store, private");
    newRes.headers.append("Vary", "Cookie");
    return newRes;
  }
  return res;
}

export { GET, POST };
