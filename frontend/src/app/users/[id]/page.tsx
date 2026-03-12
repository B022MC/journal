import Link from "next/link";
import { Container } from "@/components/layout/container";
import { AchievementBadges } from "@/components/people/achievement-badges";
import { PaperCard } from "@/components/papers/paper-card";
import { PageEmptyState } from "@/components/states/page-empty-state";
import { PageErrorState } from "@/components/states/page-error-state";
import { formatRoleLabel, formatUnixDate } from "@/lib/journal/presenters";
import { getUserPapers, getUserProfile } from "@/lib/journal/server";

export const dynamic = "force-dynamic";

export default async function PublicUserPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const [userResult, papersResult] = await Promise.all([
    getUserProfile(id),
    getUserPapers(id),
  ]);

  if (!userResult.ok) {
    return (
      <div className="py-10 sm:py-12">
        <Container>
          <PageErrorState
            title="The public profile could not be loaded."
            detail={userResult.error.detail}
            actionHref="/papers"
            actionLabel="Back to archive"
          />
        </Container>
      </div>
    );
  }

  const user = userResult.data;
  const displayName = user.nickname || user.username;

  return (
    <div className="py-10 sm:py-12">
      <Container className="space-y-6">
        <section className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
          <div className="rounded-[2rem] border border-border/80 bg-card/90 p-6 shadow-[0_30px_80px_rgba(23,20,17,0.08)] sm:p-8">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Public profile
            </p>
            <h1 className="mt-4 font-serif text-4xl tracking-tight text-foreground sm:text-5xl">
              {displayName}
            </h1>
            <p className="mt-4 max-w-3xl text-sm leading-7 text-muted-foreground sm:text-base">
              This route exposes only public identity, badges, and authored
              papers. Private actions stay in `/me`, which keeps the permission
              boundary explicit.
            </p>
            <div className="mt-6 flex flex-wrap gap-3">
              <Link
                href="/papers"
                className="inline-flex rounded-full border border-border/80 bg-background/75 px-5 py-3 text-sm font-medium text-foreground"
              >
                Back to archive
              </Link>
              <Link
                href={`/login?returnTo=/users/${id}`}
                className="inline-flex rounded-full border border-border/80 bg-background/75 px-5 py-3 text-sm font-medium text-foreground"
              >
                Sign in
              </Link>
            </div>
          </div>

          <aside className="rounded-[2rem] border border-border/80 bg-secondary/75 p-6 sm:p-7">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Public metrics
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
          </aside>
        </section>

        <section className="space-y-4">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Badges
            </p>
            <h2 className="mt-2 font-serif text-3xl tracking-tight text-foreground">
              Public unlocks
            </h2>
          </div>
          <AchievementBadges
            badges={user.achievements}
            emptyDetail="Public profiles keep the empty badge rail visible instead of silently hiding it."
          />
        </section>

        <section className="space-y-4">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Authored papers
            </p>
            <h2 className="mt-2 font-serif text-3xl tracking-tight text-foreground">
              Public archive footprint
            </h2>
          </div>
          {!papersResult.ok ? (
            <PageErrorState
              title="The public paper list could not be loaded."
              detail={papersResult.error.detail}
              actionHref="/papers"
              actionLabel="Back to archive"
            />
          ) : papersResult.data.items.length === 0 ? (
            <PageEmptyState
              title="No public papers yet."
              detail="This public profile still renders a stable empty state even when the author has not published into the archive."
              actionHref="/submit"
              actionLabel="Open submit route"
            />
          ) : (
            <div className="grid gap-4 xl:grid-cols-2">
              {papersResult.data.items.map((paper) => (
                <PaperCard key={paper.id} paper={paper} compact />
              ))}
            </div>
          )}
        </section>
      </Container>
    </div>
  );
}
