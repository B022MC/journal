import "server-only";
import { cookies } from "next/headers";
import { apiFetch, type ApiRequestOptions, type ApiResult } from "@/lib/api/client";
import { getApiRuntimeSnapshot, type AuthStrategy } from "@/lib/api/runtime";

export const AUTH_SESSION_COOKIE = "shit_journal_session";

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

export async function apiFetchWithSession<T>(
  path: string,
  options: ApiRequestOptions = {},
): Promise<ApiResult<T>> {
  const session = await getServerAuthSession();

  return apiFetch<T>(path, {
    ...options,
    access: options.access ?? "optional",
    token: options.token ?? session.token,
  });
}
