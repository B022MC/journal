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
