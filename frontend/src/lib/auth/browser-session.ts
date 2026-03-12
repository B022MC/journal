export const BROWSER_TOKEN_STORAGE_KEY = "journal.browser_token";

export function readBrowserToken() {
  if (typeof window === "undefined") {
    return null;
  }

  return window.localStorage.getItem(BROWSER_TOKEN_STORAGE_KEY);
}

export function writeBrowserToken(token: string) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(BROWSER_TOKEN_STORAGE_KEY, token);
}

export function clearBrowserToken() {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.removeItem(BROWSER_TOKEN_STORAGE_KEY);
}

type SessionBridgeResult =
  | { ok: true }
  | { ok: false; message: string };

async function readBridgeMessage(response: Response) {
  try {
    const payload = (await response.json()) as { message?: string };
    if (payload.message) {
      return payload.message;
    }
  } catch {
    // Fall through to a generic message below.
  }

  return `Session bridge returned ${response.status} ${response.statusText}.`;
}

async function syncSessionCookie(
  method: "POST" | "DELETE",
  payload?: { token: string; expireAt?: number },
): Promise<SessionBridgeResult> {
  try {
    const response = await fetch("/api/auth/session", {
      method,
      headers: payload ? { "Content-Type": "application/json" } : undefined,
      body: payload ? JSON.stringify(payload) : undefined,
      credentials: "same-origin",
    });

    if (!response.ok) {
      return {
        ok: false,
        message: await readBridgeMessage(response),
      };
    }

    return { ok: true };
  } catch (error) {
    return {
      ok: false,
      message:
        error instanceof Error
          ? error.message
          : "The browser session bridge failed before the request completed.",
    };
  }
}

export async function persistBrowserSession(token: string, expireAt?: number) {
  const result = await syncSessionCookie("POST", { token, expireAt });
  if (!result.ok) {
    return result;
  }

  writeBrowserToken(token);
  return result;
}

export async function clearBrowserSession() {
  const result = await syncSessionCookie("DELETE");
  if (!result.ok) {
    return result;
  }

  clearBrowserToken();
  return result;
}
