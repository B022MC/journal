export type ApiRuntimeSource = "env" | "same-origin";
export type AuthStrategy = "cookie-bridge" | "browser-token";

export interface ApiRuntimeSnapshot {
  origin: string | null;
  prefix: string;
  baseUrl: string;
  source: ApiRuntimeSource;
  authStrategy: AuthStrategy;
  diagnostics: string[];
}

const DEFAULT_API_PREFIX = "/api/v1";
const DEFAULT_AUTH_STRATEGY: AuthStrategy = "cookie-bridge";

function readEnv(name: string) {
  const value = process.env[name];

  if (!value) {
    return null;
  }

  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : null;
}

function normalizePrefix(prefix: string | null) {
  if (!prefix) {
    return DEFAULT_API_PREFIX;
  }

  return prefix.startsWith("/") ? prefix : `/${prefix}`;
}

function normalizeOrigin(origin: string | null) {
  if (!origin) {
    return null;
  }

  return origin.endsWith("/") ? origin.slice(0, -1) : origin;
}

function resolveAuthStrategy(value: string | null): AuthStrategy {
  if (value === "browser-token") {
    return "browser-token";
  }

  return DEFAULT_AUTH_STRATEGY;
}

export function getApiRuntimeSnapshot(): ApiRuntimeSnapshot {
  const origin = normalizeOrigin(
    readEnv("JOURNAL_API_ORIGIN") ?? readEnv("NEXT_PUBLIC_JOURNAL_API_ORIGIN"),
  );
  const prefix = normalizePrefix(
    readEnv("JOURNAL_API_PREFIX") ?? readEnv("NEXT_PUBLIC_JOURNAL_API_PREFIX"),
  );
  const authStrategy = resolveAuthStrategy(
    readEnv("JOURNAL_AUTH_STRATEGY") ??
      readEnv("NEXT_PUBLIC_JOURNAL_AUTH_STRATEGY"),
  );
  const diagnostics: string[] = [];

  if (!origin) {
    diagnostics.push(
      "JOURNAL_API_ORIGIN is unset, so requests currently fall back to same-origin /api/v1.",
    );
  }

  diagnostics.push(
    authStrategy === "cookie-bridge"
      ? "Shared auth entry keeps a migration slot for an httpOnly cookie bridge."
      : "Auth still runs in browser-token compatibility mode until the cookie bridge is ready.",
  );

  return {
    origin,
    prefix,
    baseUrl: origin ? `${origin}${prefix}` : prefix,
    source: origin ? "env" : "same-origin",
    authStrategy,
    diagnostics,
  };
}

export function resolveApiUrl(path: string) {
  if (/^https?:\/\//.test(path)) {
    return path;
  }

  const runtime = getApiRuntimeSnapshot();
  const normalizedPath = path.startsWith("/api/")
    ? path
    : `${runtime.prefix}${path.startsWith("/") ? path : `/${path}`}`;

  return runtime.origin ? `${runtime.origin}${normalizedPath}` : normalizedPath;
}
