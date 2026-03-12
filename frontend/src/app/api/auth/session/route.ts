import { NextRequest, NextResponse } from "next/server";
import { AUTH_SESSION_COOKIE } from "@/lib/auth/session-cookie";

const DEFAULT_SESSION_MAX_AGE = 60 * 60 * 24 * 7;

function sessionCookieOptions(request: NextRequest) {
  return {
    httpOnly: true,
    sameSite: "lax" as const,
    secure: request.nextUrl.protocol === "https:",
    path: "/",
  };
}

export async function POST(request: NextRequest) {
  let payload: { token?: unknown; expireAt?: unknown };

  try {
    payload = (await request.json()) as { token?: unknown; expireAt?: unknown };
  } catch {
    return NextResponse.json(
      { message: "Session bridge expects a JSON body." },
      { status: 400 },
    );
  }

  const token = typeof payload.token === "string" ? payload.token.trim() : "";
  if (!token) {
    return NextResponse.json(
      { message: "Session bridge requires a non-empty token." },
      { status: 400 },
    );
  }

  const expireAt =
    typeof payload.expireAt === "number" && Number.isFinite(payload.expireAt)
      ? Math.floor(payload.expireAt)
      : null;
  const now = Math.floor(Date.now() / 1000);
  const maxAge =
    expireAt && expireAt > now
      ? Math.max(60, expireAt - now)
      : DEFAULT_SESSION_MAX_AGE;

  const response = NextResponse.json({ success: true });
  response.cookies.set(AUTH_SESSION_COOKIE, token, {
    ...sessionCookieOptions(request),
    maxAge,
  });
  return response;
}

export async function DELETE(request: NextRequest) {
  const response = NextResponse.json({ success: true });
  response.cookies.set(AUTH_SESSION_COOKIE, "", {
    ...sessionCookieOptions(request),
    expires: new Date(0),
    maxAge: 0,
  });
  return response;
}
