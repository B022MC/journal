export type HomeReleaseVariant = "main" | "roadmap";
export type SearchReleaseEngine = "fulltext" | "hybrid" | "auto";

export interface SiteReleaseFlags {
  homeVariant: HomeReleaseVariant;
  searchDefaultEngine: SearchReleaseEngine;
}

function readEnv(name: string) {
  const value = process.env[name];
  if (!value) {
    return null;
  }

  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : null;
}

function normalizeHomeVariant(value: string | null): HomeReleaseVariant {
  return value === "roadmap" ? "roadmap" : "main";
}

function normalizeSearchEngine(value: string | null): SearchReleaseEngine {
  if (value === "hybrid" || value === "auto") {
    return value;
  }

  return "fulltext";
}

export function getSiteReleaseFlags(): SiteReleaseFlags {
  return {
    homeVariant: normalizeHomeVariant(
      readEnv("JOURNAL_HOME_VARIANT") ??
        readEnv("NEXT_PUBLIC_JOURNAL_HOME_VARIANT"),
    ),
    searchDefaultEngine: normalizeSearchEngine(
      readEnv("JOURNAL_SEARCH_RELEASE_ENGINE") ??
        readEnv("NEXT_PUBLIC_JOURNAL_SEARCH_RELEASE_ENGINE"),
    ),
  };
}
