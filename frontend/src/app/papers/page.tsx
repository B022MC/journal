import Link from "next/link";
import { Container } from "@/components/layout/container";
import { PaperCard } from "@/components/papers/paper-card";
import { PageEmptyState } from "@/components/states/page-empty-state";
import { PageErrorState } from "@/components/states/page-error-state";
import type { ListPapersResponse, SearchPapersResponse } from "@/lib/journal/contracts";
import { getSiteReleaseFlags, type SearchReleaseEngine } from "@/lib/release/flags";
import { parseSearchParam } from "@/lib/journal/presenters";
import { listPapers, searchPapers } from "@/lib/journal/server";

export const dynamic = "force-dynamic";

const papersPageSize = 12;

export default async function PapersPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const params = await searchParams;
  const releaseFlags = getSiteReleaseFlags();
  const query = parseSearchParam(params.query);
  const browsingMode = query.length === 0;
  const zone = browsingMode ? parseSearchParam(params.zone) : "";
  const discipline = parseSearchParam(params.discipline);
  const sort = parseSearchParam(params.sort, browsingMode ? "newest" : "relevance");
  const requestedEngine = parseEngineParam(params.engine);
  const effectiveEngine =
    requestedEngine === "auto" ? releaseFlags.searchDefaultEngine : requestedEngine;
  const shadowCompare = parseBooleanParam(params.shadow_compare);
  const page = parsePositiveIntParam(params.page, 1);

  const result = browsingMode
    ? await listPapers({
        discipline,
        page,
        pageSize: papersPageSize,
        sort,
        zone,
      })
    : await searchPapers({
        discipline,
        engine: requestedEngine,
        page,
        pageSize: papersPageSize,
        query,
        shadowCompare,
        sort,
        suggestionLimit: 6,
      });

  const requestEngineLabel = formatRequestedEngineLabel(
    requestedEngine,
    effectiveEngine,
  );
  const breadcrumbs = [
    query && `Query: ${query}`,
    zone && `Zone: ${zone}`,
    discipline && `Discipline: ${discipline}`,
    !browsingMode && `Request: ${requestEngineLabel}`,
    !browsingMode && shadowCompare && "Shadow compare on",
    page > 1 && `Page: ${page}`,
  ].filter((breadcrumb): breadcrumb is string => Boolean(breadcrumb));

  const searchData = result.ok && isSearchResponse(result.data) ? result.data : null;
  const suggestions = searchData?.suggestions ?? [];
  const totalPages = result.ok
    ? Math.max(1, Math.ceil(result.data.total / papersPageSize))
    : 1;
  const previousPageHref =
    page > 1
      ? buildPapersHref({
          query,
          zone,
          discipline,
          sort,
          engine: requestedEngine,
          shadowCompare: !browsingMode && shadowCompare,
          page: page - 1,
        })
      : null;
  const nextPageHref =
    result.ok && page < totalPages
      ? buildPapersHref({
          query,
          zone,
          discipline,
          sort,
          engine: requestedEngine,
          shadowCompare: !browsingMode && shadowCompare,
          page: page + 1,
        })
      : null;

  return (
    <div className="py-10 sm:py-12">
      <Container className="grid gap-6 lg:grid-cols-[320px_minmax(0,1fr)]">
        <aside className="space-y-4">
          <div className="rounded-[1.8rem] border border-border/80 bg-card/75 p-5">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Search Console
            </p>
            <details className="mt-4 rounded-[1.2rem] border border-border/70 bg-background/70 p-4 lg:open" open>
              <summary className="cursor-pointer text-sm font-medium text-foreground lg:pointer-events-none">
                Query, ranking, and fallback controls
              </summary>
              <form className="mt-4 space-y-4" method="get">
                <label className="block text-sm text-muted-foreground">
                  Query
                  <input
                    name="query"
                    defaultValue={query}
                    className="mt-2 w-full rounded-[1.2rem] border border-border/80 bg-card px-3 py-3 text-foreground"
                    placeholder="Search title, abstract, or keywords"
                  />
                </label>
                <label className="block text-sm text-muted-foreground">
                  Zone
                  <select
                    name="zone"
                    defaultValue={zone}
                    disabled={!browsingMode}
                    className="mt-2 w-full rounded-[1.2rem] border border-border/80 bg-card px-3 py-3 text-foreground disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <option value="">All zones</option>
                    <option value="latrine">Latrine</option>
                    <option value="septic_tank">Septic Tank</option>
                    <option value="stone">Stone</option>
                    <option value="sediment">Sediment</option>
                  </select>
                </label>
                <label className="block text-sm text-muted-foreground">
                  Discipline
                  <input
                    name="discipline"
                    defaultValue={discipline}
                    className="mt-2 w-full rounded-[1.2rem] border border-border/80 bg-card px-3 py-3 text-foreground"
                    placeholder="Biology, CS, sociology..."
                  />
                </label>
                <label className="block text-sm text-muted-foreground">
                  Sort
                  <select
                    name="sort"
                    defaultValue={sort}
                    className="mt-2 w-full rounded-[1.2rem] border border-border/80 bg-card px-3 py-3 text-foreground"
                  >
                    {!browsingMode ? <option value="relevance">Relevance</option> : null}
                    <option value="newest">Newest</option>
                    <option value="quality">Quality</option>
                  </select>
                </label>
                <label className="block text-sm text-muted-foreground">
                  Search engine
                  <select
                    name="engine"
                    defaultValue={requestedEngine}
                    disabled={browsingMode}
                    className="mt-2 w-full rounded-[1.2rem] border border-border/80 bg-card px-3 py-3 text-foreground disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <option value="auto">Server default</option>
                    <option value="hybrid">Hybrid index</option>
                    <option value="fulltext">MySQL FULLTEXT</option>
                  </select>
                </label>
                <label className="flex items-start gap-3 rounded-[1.2rem] border border-border/70 bg-card/70 p-3 text-sm text-muted-foreground">
                  <input
                    type="checkbox"
                    name="shadow_compare"
                    defaultChecked={shadowCompare}
                    disabled={browsingMode}
                    className="mt-1 h-4 w-4 rounded border-border/80"
                  />
                  <span>
                    Shadow compare new search against FULLTEXT while keeping the visible response on the safe path.
                  </span>
                </label>
                <p className="rounded-[1rem] border border-border/60 bg-background/60 px-3 py-2 text-xs leading-6 text-muted-foreground">
                  Release default engine:{" "}
                  <span className="font-medium text-foreground">
                    {releaseFlags.searchDefaultEngine}
                  </span>
                  . Current request follows{" "}
                  <span className="font-medium text-foreground">
                    {requestEngineLabel}
                  </span>
                  . Change `JOURNAL_SEARCH_RELEASE_ENGINE` to switch or roll
                  back default traffic without editing route code.
                </p>
                {!browsingMode && requestedEngine !== "auto" ? (
                  <p className="rounded-[1rem] border border-border/60 bg-background/60 px-3 py-2 text-xs text-muted-foreground">
                    This URL pins {requestedEngine}. Switch Search engine back to
                    Server default to follow the current release flag without
                    changing the validation route.
                  </p>
                ) : null}
                {!browsingMode ? null : (
                  <p className="rounded-[1rem] border border-border/60 bg-background/60 px-3 py-2 text-xs text-muted-foreground">
                    Fallback and shadow controls activate after a query is present.
                    Browse mode keeps zone filtering on the list endpoint and uses
                    page, zone, discipline, and sort as the only active URL state.
                  </p>
                )}
                <div className="flex flex-wrap gap-3">
                  <button
                    type="submit"
                    className="inline-flex rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
                  >
                    {browsingMode ? "Browse papers" : "Run search"}
                  </button>
                  <Link
                    href="/papers"
                    className="inline-flex rounded-full border border-border/80 bg-background/75 px-4 py-2 text-sm font-medium text-foreground"
                  >
                    Reset
                  </Link>
                </div>
              </form>
            </details>
          </div>

          {suggestions.length > 0 ? (
            <div className="rounded-[1.8rem] border border-border/80 bg-card/75 p-5">
              <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
                Suggestion Chips
              </p>
              <div className="mt-4 flex flex-wrap gap-2">
                {suggestions.map((suggestion) => (
                  <Link
                    key={suggestion}
                    href={buildPapersHref({
                      discipline,
                      engine: requestedEngine,
                      query: suggestion,
                      shadowCompare,
                      sort,
                      page: 1,
                    })}
                    className="rounded-full border border-border/80 bg-background/70 px-3 py-2 text-xs uppercase tracking-[0.18em] text-foreground"
                  >
                    {suggestion}
                  </Link>
                ))}
              </div>
            </div>
          ) : null}
        </aside>

        <section className="space-y-5">
          <div className="rounded-[1.8rem] border border-border/80 bg-card/75 p-5">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Paper Index
            </p>
            <h1 className="mt-3 font-serif text-4xl tracking-tight text-foreground">
              {browsingMode ? "Searchable archive of the current paper flow" : "Search comparison surface for the paper archive"}
            </h1>
            <div className="mt-4 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
              <span>{result.ok ? `${result.data.total} result(s)` : "Live data unavailable"}</span>
              {breadcrumbs.map((breadcrumb) => (
                <span
                  key={breadcrumb}
                  className="rounded-full bg-background/75 px-3 py-1 text-xs uppercase tracking-[0.18em]"
                >
                  {breadcrumb}
                </span>
              ))}
            </div>

            {searchData ? (
              <div className="mt-4 space-y-3 rounded-[1.4rem] border border-border/70 bg-background/70 p-4 text-sm text-muted-foreground">
                <div className="flex flex-wrap gap-2">
                  <span className="rounded-full bg-card px-3 py-1 text-xs uppercase tracking-[0.18em] text-foreground">
                    Engine {searchData.meta.engine}
                  </span>
                  <span className="rounded-full bg-card px-3 py-1 text-xs uppercase tracking-[0.18em] text-foreground">
                    Release default {releaseFlags.searchDefaultEngine}
                  </span>
                  {searchData.meta.used_fallback ? (
                    <span className="rounded-full bg-accent px-3 py-1 text-xs uppercase tracking-[0.18em] text-foreground">
                      Fallback {searchData.meta.fallback_reason || "triggered"}
                    </span>
                  ) : null}
                  {searchData.meta.shadow_compared ? (
                    <span className="rounded-full bg-card px-3 py-1 text-xs uppercase tracking-[0.18em] text-foreground">
                      Shadow compared
                    </span>
                  ) : null}
                  <span className="rounded-full bg-card px-3 py-1 text-xs uppercase tracking-[0.18em] text-foreground">
                    Index {searchData.meta.indexed_docs} docs / {searchData.meta.indexed_terms} terms
                  </span>
                </div>
                {searchData.meta.used_fallback ? (
                  <p className="text-xs leading-6 text-muted-foreground">
                    The visible response stayed on the safe path. Use the fallback
                    reason above together with the release default label to verify
                    whether the page is following the default route or an explicit
                    validation override.
                  </p>
                ) : null}
                {searchData.meta.expanded_terms.length > 0 ? (
                  <div className="flex flex-wrap gap-2">
                    {searchData.meta.expanded_terms.map((term) => (
                      <span
                        key={term}
                        className="rounded-full border border-border/70 bg-card/70 px-3 py-1 text-xs uppercase tracking-[0.18em] text-foreground"
                      >
                        {term}
                      </span>
                    ))}
                  </div>
                ) : null}
              </div>
            ) : null}
          </div>

          {!result.ok ? (
            <PageErrorState
              detail={result.error.detail}
              actionHref={
                !browsingMode
                  ? buildPapersHref({
                      query,
                      discipline,
                      sort,
                      engine: "fulltext",
                    })
                  : "/login?returnTo=/papers"
              }
              actionLabel={!browsingMode ? "Retry with FULLTEXT" : "Check sign-in path"}
            />
          ) : result.data.items.length === 0 ? (
            <PageEmptyState
              title={!browsingMode ? "No search results matched the current query." : "No papers matched the current filters."}
              detail={!browsingMode ? "Try a broader query, switch the engine, or clear discipline filters." : "Try switching to a broader zone and discipline combination."}
              actionHref="/papers"
              actionLabel="Reset filters"
            />
          ) : (
            <div className="space-y-4">
              {result.data.items.map((paper) => (
                <PaperCard key={paper.id} paper={paper} />
              ))}
            </div>
          )}

          {result.ok && totalPages > 1 ? (
            <div className="flex flex-wrap items-center justify-between gap-3 rounded-[1.4rem] border border-border/70 bg-card/75 p-4 text-sm text-muted-foreground">
              <p>
                Page {page} of {totalPages}
              </p>
              <div className="flex flex-wrap gap-3">
                {previousPageHref ? (
                  <Link
                    href={previousPageHref}
                    className="inline-flex rounded-full border border-border/80 bg-background/75 px-4 py-2 text-sm font-medium text-foreground"
                  >
                    Previous page
                  </Link>
                ) : (
                  <span className="inline-flex cursor-not-allowed rounded-full border border-border/60 bg-background/50 px-4 py-2 text-sm font-medium opacity-50">
                    Previous page
                  </span>
                )}
                {nextPageHref ? (
                  <Link
                    href={nextPageHref}
                    className="inline-flex rounded-full border border-border/80 bg-background/75 px-4 py-2 text-sm font-medium text-foreground"
                  >
                    Next page
                  </Link>
                ) : (
                  <span className="inline-flex cursor-not-allowed rounded-full border border-border/60 bg-background/50 px-4 py-2 text-sm font-medium opacity-50">
                    Next page
                  </span>
                )}
              </div>
            </div>
          ) : null}
        </section>
      </Container>
    </div>
  );
}

function parseEngineParam(value: string | string[] | undefined): SearchReleaseEngine {
  const normalized = parseSearchParam(value, "auto").toLowerCase();
  if (normalized === "fulltext" || normalized === "hybrid" || normalized === "auto") {
    return normalized;
  }
  return "auto";
}

function parseBooleanParam(value: string | string[] | undefined) {
  const normalized = parseSearchParam(value).toLowerCase();
  return normalized === "1" || normalized === "true" || normalized === "on";
}

function parsePositiveIntParam(value: string | string[] | undefined, fallback: number) {
  const parsed = Number.parseInt(parseSearchParam(value, String(fallback)), 10);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback;
  }
  return parsed;
}

function formatRequestedEngineLabel(
  requestedEngine: SearchReleaseEngine,
  effectiveEngine: SearchReleaseEngine,
) {
  if (requestedEngine === "auto") {
    return `server default (${effectiveEngine})`;
  }
  return `${requestedEngine} override`;
}

function buildPapersHref({
  query,
  zone,
  discipline,
  sort,
  engine,
  shadowCompare,
  page,
}: {
  query?: string;
  zone?: string;
  discipline?: string;
  sort?: string;
  engine?: SearchReleaseEngine;
  shadowCompare?: boolean;
  page?: number;
}) {
  const params = new URLSearchParams();
  if (query) {
    params.set("query", query);
  } else if (zone) {
    params.set("zone", zone);
  }
  if (discipline) {
    params.set("discipline", discipline);
  }
  if (sort) {
    params.set("sort", sort);
  }
  if (query && engine && engine !== "auto") {
    params.set("engine", engine);
  }
  if (query && shadowCompare) {
    params.set("shadow_compare", "true");
  }
  if (page && page > 1) {
    params.set("page", String(page));
  }
  const search = params.toString();
  return search ? `/papers?${search}` : "/papers";
}

function isSearchResponse(data: ListPapersResponse | SearchPapersResponse): data is SearchPapersResponse {
  return "meta" in data;
}
