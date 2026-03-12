import { ApiConfigError, ApiRequestError, type ApiDiagnostic } from "@/lib/api/errors";
import { resolveApiUrl } from "@/lib/api/runtime";

type QueryValue = string | number | boolean | null | undefined;
type QueryRecord = Record<string, QueryValue | QueryValue[]>;
type JsonRecord = Record<string, unknown>;

export interface ApiSuccess<T> {
  ok: true;
  data: T;
  response: Response;
}

export interface ApiFailure {
  ok: false;
  error: ApiDiagnostic;
  response?: Response;
}

export type ApiResult<T> = ApiSuccess<T> | ApiFailure;

export interface ApiRequestOptions extends Omit<RequestInit, "body"> {
  access?: "public" | "optional" | "required";
  body?: BodyInit | JsonRecord | null;
  query?: QueryRecord;
  token?: string | null;
}

function appendQuery(url: URL, query?: QueryRecord) {
  if (!query) {
    return;
  }

  for (const [key, rawValue] of Object.entries(query)) {
    const values = Array.isArray(rawValue) ? rawValue : [rawValue];

    for (const value of values) {
      if (value === undefined || value === null || value === "") {
        continue;
      }

      url.searchParams.append(key, String(value));
    }
  }
}

function resolveRequestUrl(path: string, query?: QueryRecord) {
  const url = new URL(resolveApiUrl(path), "http://journal.local");
  appendQuery(url, query);

  if (url.origin === "http://journal.local") {
    return `${url.pathname}${url.search}`;
  }

  return url.toString();
}

function normalizeBody(body: ApiRequestOptions["body"], headers: Headers) {
  if (!body || typeof body !== "object" || body instanceof FormData || body instanceof Blob || body instanceof ArrayBuffer || body instanceof URLSearchParams || ArrayBuffer.isView(body)) {
    return body ?? undefined;
  }

  headers.set("Content-Type", "application/json");
  return JSON.stringify(body);
}

async function readFailureDetail(response: Response) {
  const contentType = response.headers.get("content-type") ?? "";

  if (contentType.includes("application/json")) {
    try {
      const payload = (await response.json()) as {
        message?: string;
        msg?: string;
        error?: string;
      };
      return payload.message ?? payload.msg ?? payload.error ?? response.statusText;
    } catch {
      return response.statusText;
    }
  }

  try {
    const text = await response.text();
    return text.trim() || response.statusText;
  } catch {
    return response.statusText;
  }
}

export async function apiFetch<T>(
  path: string,
  options: ApiRequestOptions = {},
): Promise<ApiResult<T>> {
  const {
    access = "public",
    body,
    headers: initialHeaders,
    method = "GET",
    query,
    token,
    ...rest
  } = options;

  if (access === "required" && !token) {
    const error = new ApiConfigError(
      "This request requires an authenticated session, but no token was provided.",
      "AUTH_TOKEN_MISSING",
    );

    return { ok: false, error: error.diagnostic };
  }

  const headers = new Headers(initialHeaders);
  headers.set("Accept", "application/json");

  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const requestBody = normalizeBody(body, headers);
  const url = resolveRequestUrl(path, query);

  let response: Response;

  try {
    response = await fetch(url, {
      ...rest,
      body: requestBody,
      cache: rest.cache ?? "no-store",
      headers,
      method,
    });
  } catch (error) {
    return {
      ok: false,
      error: new ApiRequestError({
        kind: "network",
        title: "API request failed",
        detail:
          error instanceof Error
            ? error.message
            : "The request failed before a response was received.",
        retryable: true,
        code: "API_NETWORK_FAILURE",
      }).diagnostic,
    };
  }

  if (!response.ok) {
    const detail = await readFailureDetail(response);

    return {
      ok: false,
      error: new ApiRequestError({
        kind: response.status === 401 ? "auth" : "http",
        title:
          response.status === 401
            ? "Authentication required"
            : "API returned an error",
        detail,
        retryable: response.status >= 500,
        status: response.status,
        code: `HTTP_${response.status}`,
      }).diagnostic,
      response,
    };
  }

  if (response.status === 204) {
    return { ok: true, data: undefined as T, response };
  }

  const contentType = response.headers.get("content-type") ?? "";

  if (contentType.includes("application/json")) {
    const data = (await response.json()) as T;
    return { ok: true, data, response };
  }

  const data = (await response.text()) as T;
  return { ok: true, data, response };
}

export function unwrapApiResult<T>(result: ApiResult<T>) {
  if (!result.ok) {
    throw new ApiRequestError(result.error);
  }

  return result.data;
}
