import Link from "next/link";
import { Container } from "@/components/layout/container";
import { PaperCard } from "@/components/papers/paper-card";
import { PageEmptyState } from "@/components/states/page-empty-state";
import { PageErrorState } from "@/components/states/page-error-state";
import { getCurrentUser, listPapers } from "@/lib/journal/server";
import { buildZoneSummary } from "@/lib/journal/presenters";

export const dynamic = "force-dynamic";

export default async function HomePage() {
  const [papersResult, userResult] = await Promise.all([
    listPapers({ pageSize: 6 }),
    getCurrentUser(),
  ]);

  const userName = userResult.ok
    ? userResult.data.nickname || userResult.data.username
    : null;

  return (
    <div className="py-10 sm:py-12">
      <Container className="grid gap-6 lg:grid-cols-[minmax(0,1.1fr)_minmax(280px,0.55fr)]">
        <section className="rounded-[2rem] border border-border/80 bg-card/90 p-6 shadow-[0_30px_80px_rgba(23,20,17,0.10)] sm:p-8">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Live Editorial Desk
          </p>
          <h1 className="mt-5 max-w-3xl font-serif text-4xl leading-none tracking-tight text-foreground sm:text-5xl lg:text-6xl">
            Reading, scoring, and governance now have a real front door.
          </h1>
          <p className="mt-5 max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg">
            The main site has moved beyond the roadmap splash page. Phase 1
            now prioritizes the browse-to-detail path, with auth and governance
            interaction islands layered on top of the server-rendered shell.
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Link
              href="/papers"
              className="inline-flex rounded-full bg-primary px-5 py-3 text-sm font-medium text-primary-foreground"
            >
              Browse Papers
            </Link>
            <Link
              href="/register?returnTo=/papers"
              className="inline-flex rounded-full border border-border/80 bg-background/75 px-5 py-3 text-sm font-medium text-foreground"
            >
              Submit Paper
            </Link>
          </div>
        </section>

        <aside className="rounded-[2rem] border border-border/80 bg-secondary/75 p-6 sm:p-7">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Session + Flow
          </p>
          <div className="mt-5 space-y-4">
            <div className="rounded-[1.4rem] border border-border/70 bg-background/75 p-4">
              <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                Current view
              </p>
              <p className="mt-2 text-lg font-semibold text-foreground">
                {userName ? `Signed in as ${userName}` : "Anonymous reader"}
              </p>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">
                Login and registration stay on dedicated pages so the browse and
                detail routes can remain server-first.
              </p>
            </div>
            <div className="rounded-[1.4rem] border border-border/70 bg-background/75 p-4">
              <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                Governance lane
              </p>
              <p className="mt-2 text-sm leading-6 text-foreground">
                Rating and report actions remain in the paper sidebar. They do
                not block reading, and they degrade cleanly when the auth token
                is missing.
              </p>
            </div>
          </div>
        </aside>
      </Container>

      <Container className="mt-6 grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
        <section className="space-y-5">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
                Phase 1 Home
              </p>
              <h2 className="mt-2 font-serif text-3xl tracking-tight text-foreground">
                Fresh papers from the archive desk
              </h2>
            </div>
            <Link
              href="/papers"
              className="text-sm font-medium text-[#8a4b2a]"
            >
              View full index
            </Link>
          </div>

          {!papersResult.ok ? (
            <PageErrorState
              detail={papersResult.error.detail}
              actionHref="/login"
              actionLabel="Check auth path"
            />
          ) : papersResult.data.items.length === 0 ? (
            <PageEmptyState
              title="No papers are visible yet."
              detail="The shell is ready, but the archive feed has not been populated. Once the API starts returning papers, this desk automatically becomes the live homepage."
              actionHref="/register"
              actionLabel="Create the first account"
            />
          ) : (
            <div className="grid gap-4 xl:grid-cols-2">
              {papersResult.data.items.map((paper) => (
                <PaperCard key={paper.id} paper={paper} compact />
              ))}
            </div>
          )}
        </section>

        <aside className="space-y-4">
          <section className="rounded-[1.8rem] border border-border/80 bg-card/75 p-5">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Zone Overview
            </p>
            <ul className="mt-4 space-y-3">
              {(papersResult.ok ? buildZoneSummary(papersResult.data.items) : []).map((zone) => (
                <li
                  key={zone.zone}
                  className="rounded-[1.2rem] border border-border/70 bg-background/70 px-4 py-3"
                >
                  <p className="text-sm font-semibold text-foreground">
                    {zone.zone || "Unclassified"}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {zone.count} item(s) · latest:
                    {" "}
                    {zone.latestTitle ?? "n/a"}
                  </p>
                </li>
              ))}
              {papersResult.ok && papersResult.data.items.length === 0 ? (
                <li className="rounded-[1.2rem] border border-dashed border-border/70 px-4 py-3 text-sm text-muted-foreground">
                  Zone metrics appear when live papers are available.
                </li>
              ) : null}
            </ul>
          </section>

          <section className="rounded-[1.8rem] border border-border/80 bg-card/75 p-5">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Governance Primer
            </p>
            <ul className="mt-4 space-y-3 text-sm leading-6 text-muted-foreground">
              <li>Scores and flags stay visible next to reading, not above it.</li>
              <li>Phase 1 keeps data fetching on the server, while auth actions stay in client islands.</li>
              <li>When the API is unavailable, pages surface a stable error card instead of reverting to the old roadmap screen.</li>
            </ul>
          </section>
        </aside>
      </Container>
    </div>
  );
}
