import Link from "next/link";
import { Container } from "@/components/layout/container";
import { LogoutButton } from "@/components/auth/logout-button";
import { AchievementBadges } from "@/components/people/achievement-badges";
import { PaperCard } from "@/components/papers/paper-card";
import { PageEmptyState } from "@/components/states/page-empty-state";
import { PageErrorState } from "@/components/states/page-error-state";
import { requireCurrentUser } from "@/lib/journal/protected";
import {
  getCurrentUserPapers,
  getCurrentUserRatings,
} from "@/lib/journal/server";
import {
  formatRoleLabel,
  formatUnixDate,
  parseSearchParam,
} from "@/lib/journal/presenters";

export const dynamic = "force-dynamic";

export default async function MePage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const currentUser = await requireCurrentUser("/me");
  if (!currentUser.ok) {
    return (
      <div className="py-10 sm:py-12">
        <Container>
          <PageErrorState
            title="The personal workspace could not be loaded."
            detail={currentUser.error.detail}
            actionHref="/papers"
            actionLabel="Back to archive"
          />
        </Container>
      </div>
    );
  }

  const params = await searchParams;
  const submittedId = parseSearchParam(params.submitted);
  const [papersResult, ratingsResult] = await Promise.all([
    getCurrentUserPapers(),
    getCurrentUserRatings(),
  ]);

  const user = currentUser.user;
  const displayName = user.nickname || user.username;

  return (
    <div className="py-10 sm:py-12">
      <Container className="space-y-6">
        {submittedId ? (
          <section className="rounded-[1.6rem] border border-[#426b54]/20 bg-[#426b54]/8 p-5">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-[#426b54]">
              Submission queued
            </p>
            <p className="mt-3 text-sm leading-7 text-foreground">
              Paper #{submittedId} was handed back to the workspace. Review the
              current state here or jump to the public detail route once the API
              confirms it.
            </p>
            <Link
              href={`/papers/${submittedId}`}
              className="mt-4 inline-flex rounded-full border border-border/80 bg-background/80 px-4 py-2 text-sm text-foreground"
            >
              Open detail route
            </Link>
          </section>
        ) : null}

        <section className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
          <div className="rounded-[2rem] border border-border/80 bg-card/90 p-6 shadow-[0_30px_80px_rgba(23,20,17,0.08)] sm:p-8">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Workspace
            </p>
            <h1 className="mt-4 font-serif text-4xl tracking-tight text-foreground sm:text-5xl">
              {displayName}
            </h1>
            <p className="mt-4 max-w-3xl text-sm leading-7 text-muted-foreground sm:text-base">
              `/me` stays protected and server-rendered. Refresh recovery comes
              from the cookie bridge, while rating, reporting, and submission
              actions still reuse the existing browser token compatibility path.
            </p>

            <div className="mt-6 flex flex-wrap gap-3">
              <Link
                href="#overview"
                className="inline-flex rounded-full border border-border/80 bg-background/75 px-4 py-2 text-sm text-foreground"
              >
                Overview
              </Link>
              <Link
                href="#papers"
                className="inline-flex rounded-full border border-border/80 bg-background/75 px-4 py-2 text-sm text-foreground"
              >
                My Papers
              </Link>
              <Link
                href="#ratings"
                className="inline-flex rounded-full border border-border/80 bg-background/75 px-4 py-2 text-sm text-foreground"
              >
                My Ratings
              </Link>
              <Link
                href="#governance"
                className="inline-flex rounded-full border border-border/80 bg-background/75 px-4 py-2 text-sm text-foreground"
              >
                Governance
              </Link>
            </div>
          </div>

          <aside className="rounded-[2rem] border border-border/80 bg-secondary/75 p-6 sm:p-7">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Account summary
            </p>
            <dl className="mt-5 grid gap-3">
              <div className="rounded-[1.2rem] border border-border/70 bg-background/75 p-4">
                <dt className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                  Role
                </dt>
                <dd className="mt-2 text-lg font-semibold text-foreground">
                  {formatRoleLabel(user.role)}
                </dd>
              </div>
              <div className="rounded-[1.2rem] border border-border/70 bg-background/75 p-4">
                <dt className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                  Contribution
                </dt>
                <dd className="mt-2 text-lg font-semibold text-foreground">
                  {user.contribution_score || "0.00"}
                </dd>
              </div>
              <div className="rounded-[1.2rem] border border-border/70 bg-background/75 p-4">
                <dt className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                  Joined
                </dt>
                <dd className="mt-2 text-sm text-foreground">
                  {formatUnixDate(user.created_at)}
                </dd>
              </div>
            </dl>
            <div className="mt-5 flex flex-wrap gap-3">
              <Link
                href={`/users/${user.id}`}
                className="inline-flex rounded-full border border-border/80 bg-background/80 px-4 py-2 text-sm text-foreground"
              >
                View public profile
              </Link>
              <LogoutButton returnTo="/" />
            </div>
          </aside>
        </section>

        <section id="overview" className="space-y-4">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Overview
            </p>
            <h2 className="mt-2 font-serif text-3xl tracking-tight text-foreground">
              Badges and role stay stable even when data is sparse.
            </h2>
          </div>
          <AchievementBadges
            badges={user.achievements}
            emptyDetail="Your workspace keeps the badge rail visible even before the first unlock, so the empty state is explicit rather than missing."
          />
        </section>

        <section id="papers" className="space-y-4">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
                My Papers
              </p>
              <h2 className="mt-2 font-serif text-3xl tracking-tight text-foreground">
                Recent submissions
              </h2>
            </div>
            <Link href="/submit" className="text-sm font-medium text-[#8a4b2a]">
              New submission
            </Link>
          </div>

          {!papersResult.ok ? (
            <PageErrorState
              title="The workspace could not load your paper list."
              detail={papersResult.error.detail}
              actionHref="/submit"
              actionLabel="Open submit studio"
            />
          ) : papersResult.data.items.length === 0 ? (
            <PageEmptyState
              title="No personal submissions yet."
              detail="The workspace remains useful before the first paper lands. Use the protected submit route to close the publish loop."
              actionHref="/submit"
              actionLabel="Submit your first paper"
            />
          ) : (
            <div className="grid gap-4 xl:grid-cols-2">
              {papersResult.data.items.map((paper) => (
                <PaperCard key={paper.id} paper={paper} compact />
              ))}
            </div>
          )}
        </section>

        <section id="ratings" className="space-y-4">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              My Ratings
            </p>
            <h2 className="mt-2 font-serif text-3xl tracking-tight text-foreground">
              Recent governance actions
            </h2>
          </div>

          {!ratingsResult.ok ? (
            <PageErrorState
              title="The workspace could not load your rating history."
              detail={ratingsResult.error.detail}
              actionHref="/papers"
              actionLabel="Return to archive"
            />
          ) : ratingsResult.data.items.length === 0 ? (
            <PageEmptyState
              title="No ratings recorded yet."
              detail="Browse a paper and submit a score to make the interaction-to-workspace loop visible here."
              actionHref="/papers"
              actionLabel="Browse papers"
            />
          ) : (
            <ul className="grid gap-4 lg:grid-cols-2">
              {ratingsResult.data.items.map((rating) => (
                <li
                  key={rating.id}
                  className="rounded-[1.6rem] border border-border/80 bg-card/80 p-5"
                >
                  <div className="flex items-center justify-between gap-3">
                    <p className="text-sm font-semibold text-foreground">
                      <Link
                        href={`/papers/${rating.paper_id}`}
                        className="transition-colors hover:text-[#8a4b2a]"
                      >
                        Paper #{rating.paper_id}
                      </Link>
                    </p>
                    <span className="rounded-full bg-background/80 px-3 py-1 text-xs uppercase tracking-[0.18em] text-muted-foreground">
                      {rating.score}/10
                    </span>
                  </div>
                  <p className="mt-2 text-xs uppercase tracking-[0.18em] text-muted-foreground">
                    {formatUnixDate(rating.created_at)}
                  </p>
                  <p className="mt-3 text-sm leading-6 text-muted-foreground">
                    {rating.comment || "No comment supplied."}
                  </p>
                </li>
              ))}
            </ul>
          )}
        </section>

        <section id="governance" className="space-y-4">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Governance
            </p>
            <h2 className="mt-2 font-serif text-3xl tracking-tight text-foreground">
              Session and permission edges
            </h2>
          </div>
          <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_320px]">
            <article className="rounded-[1.8rem] border border-border/80 bg-card/80 p-5">
              <ul className="space-y-3 text-sm leading-7 text-muted-foreground">
                <li>Protected routes redirect through `/login` with a stable `returnTo` when the shared session is missing.</li>
                <li>Refresh recovery depends on the httpOnly cookie bridge, while form submissions still reuse the existing browser token path.</li>
                <li>Submission failures and validation errors keep local draft state intact instead of clearing the workspace.</li>
              </ul>
            </article>
            <aside className="rounded-[1.8rem] border border-border/80 bg-secondary/70 p-5">
              <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
                Admin permissions
              </p>
              {user.admin_permissions.length > 0 ? (
                <ul className="mt-4 space-y-2 text-sm leading-6 text-foreground">
                  {user.admin_permissions.map((permission) => (
                    <li
                      key={permission}
                      className="rounded-[1rem] border border-border/70 bg-background/75 px-3 py-2"
                    >
                      {permission}
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="mt-4 text-sm leading-6 text-muted-foreground">
                  No admin permissions are attached to this account.
                </p>
              )}
            </aside>
          </div>
        </section>
      </Container>
    </div>
  );
}
