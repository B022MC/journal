import Link from "next/link";
import { Container } from "@/components/layout/container";
import { PaperCard } from "@/components/papers/paper-card";
import { PageEmptyState } from "@/components/states/page-empty-state";
import { PageErrorState } from "@/components/states/page-error-state";
import { parseSearchParam } from "@/lib/journal/presenters";
import { listPapers } from "@/lib/journal/server";

export const dynamic = "force-dynamic";

export default async function PapersPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const params = await searchParams;
  const query = parseSearchParam(params.query);
  const zone = parseSearchParam(params.zone);
  const discipline = parseSearchParam(params.discipline);
  const sort = parseSearchParam(params.sort, "newest");
  const page = Number.parseInt(parseSearchParam(params.page, "1"), 10) || 1;
  const result = await listPapers({
    discipline,
    page,
    pageSize: 12,
    query,
    sort,
    zone,
  });

  const breadcrumbs = [query && `Query: ${query}`, zone && `Zone: ${zone}`, discipline && `Discipline: ${discipline}`].filter(Boolean);

  return (
    <div className="py-10 sm:py-12">
      <Container className="grid gap-6 lg:grid-cols-[280px_minmax(0,1fr)]">
        <aside className="space-y-4">
          <div className="rounded-[1.8rem] border border-border/80 bg-card/75 p-5">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Archive Filters
            </p>
            <details className="mt-4 rounded-[1.2rem] border border-border/70 bg-background/70 p-4 lg:open" open>
              <summary className="cursor-pointer text-sm font-medium text-foreground lg:pointer-events-none">
                Search, sort, and narrow the reading archive
              </summary>
              <form className="mt-4 space-y-4" method="get">
                <label className="block text-sm text-muted-foreground">
                  Query
                  <input
                    name="query"
                    defaultValue={query}
                    className="mt-2 w-full rounded-[1.2rem] border border-border/80 bg-card px-3 py-3 text-foreground"
                    placeholder="Search title or abstract"
                  />
                </label>
                <label className="block text-sm text-muted-foreground">
                  Zone
                  <select
                    name="zone"
                    defaultValue={zone}
                    className="mt-2 w-full rounded-[1.2rem] border border-border/80 bg-card px-3 py-3 text-foreground"
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
                    placeholder="Biology, CS, sociology…"
                  />
                </label>
                <label className="block text-sm text-muted-foreground">
                  Sort
                  <select
                    name="sort"
                    defaultValue={sort}
                    className="mt-2 w-full rounded-[1.2rem] border border-border/80 bg-card px-3 py-3 text-foreground"
                  >
                    <option value="newest">Newest</option>
                    <option value="highest_rated">Highest rated</option>
                    <option value="most_viewed">Most viewed</option>
                  </select>
                </label>
                <div className="flex flex-wrap gap-3">
                  <button
                    type="submit"
                    className="inline-flex rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
                  >
                    Apply filters
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
        </aside>

        <section className="space-y-5">
          <div className="rounded-[1.8rem] border border-border/80 bg-card/75 p-5">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Paper Index
            </p>
            <h1 className="mt-3 font-serif text-4xl tracking-tight text-foreground">
              Searchable archive of the current paper flow
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
          </div>

          {!result.ok ? (
            <PageErrorState
              detail={result.error.detail}
              actionHref="/login?returnTo=/papers"
              actionLabel="Check sign-in path"
            />
          ) : result.data.items.length === 0 ? (
            <PageEmptyState
              title="No papers matched the current filters."
              detail="Try clearing the query or switching to a broader zone and discipline combination."
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
        </section>
      </Container>
    </div>
  );
}
