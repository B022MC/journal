import "server-only";
import { cookies, headers } from "next/headers";
import { apiFetch, type ApiRequestOptions, type ApiResult } from "@/lib/api/client";
import { AUTH_SESSION_COOKIE } from "@/lib/auth/session-cookie";
import {
  getApiRuntimeSnapshot,
  resolveApiUrl,
  type AuthStrategy,
} from "@/lib/api/runtime";

export { AUTH_SESSION_COOKIE } from "@/lib/auth/session-cookie";

export interface ServerAuthSession {
  token: string | null;
  authStrategy: AuthStrategy;
  source: "cookie" | "migration-slot";
}

export async function getServerAuthSession(): Promise<ServerAuthSession> {
  const runtime = getApiRuntimeSnapshot();
  const cookieStore = await cookies();
  const token = cookieStore.get(AUTH_SESSION_COOKIE)?.value ?? null;

  return {
    token,
    authStrategy: runtime.authStrategy,
    source: token ? "cookie" : "migration-slot",
  };
}

async function resolveRequestOrigin() {
  const runtime = getApiRuntimeSnapshot();

  if (runtime.origin) {
    return null;
  }

  const headerStore = await headers();
  const host =
    headerStore.get("x-forwarded-host") ?? headerStore.get("host") ?? null;

  if (!host) {
    return null;
  }

  const protocol = headerStore.get("x-forwarded-proto") ?? "http";
  return `${protocol}://${host}`;
}

function resolveServerApiPath(path: string, origin: string | null) {
  if (!origin || /^https?:\/\//.test(path)) {
    return path;
  }

  const resolved = resolveApiUrl(path);

  if (/^https?:\/\//.test(resolved)) {
    return resolved;
  }

  return `${origin}${resolved}`;
}

export async function apiFetchWithSession<T>(
  path: string,
  options: ApiRequestOptions = {},
): Promise<ApiResult<T>> {
  const session = await getServerAuthSession();
  const origin = await resolveRequestOrigin();
  const requestPath = resolveServerApiPath(path, origin);

  return apiFetch<T>(requestPath, {
    ...options,
    access: options.access ?? "optional",
    token: options.token ?? session.token,
  });
}
