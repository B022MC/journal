import Link from "next/link";
import { Container } from "@/components/layout/container";
import { FlagPaperForm } from "@/components/papers/flag-paper-form";
import { RatingComposer } from "@/components/papers/rating-composer";
import { PageErrorState } from "@/components/states/page-error-state";
import {
  formatCompactNumber,
  formatScore,
  formatUnixDate,
  getZoneTone,
  splitKeywords,
} from "@/lib/journal/presenters";
import { getPaper, getPaperRatings } from "@/lib/journal/server";

export const dynamic = "force-dynamic";

export default async function PaperDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const [paperResult, ratingsResult] = await Promise.all([
    getPaper(id),
    getPaperRatings(id),
  ]);

  if (!paperResult.ok) {
    return (
      <div className="py-10 sm:py-12">
        <Container>
          <PageErrorState
            title="The paper detail could not be loaded."
            detail={paperResult.error.detail}
            actionHref="/papers"
            actionLabel="Back to archive"
          />
        </Container>
      </div>
    );
  }

  const paper = paperResult.data;
  const zoneTone = getZoneTone(paper.zone);
  const contentBlocks = (paper.content || "").split(/\n{2,}/).filter(Boolean);

  return (
    <div className="py-10 sm:py-12">
      <Container className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
        <article className="rounded-[2rem] border border-border/80 bg-card/85 p-6 shadow-[0_30px_80px_rgba(23,20,17,0.08)] sm:p-8">
          <Link
            href="/papers"
            className="text-sm font-medium text-[#8a4b2a]"
          >
            ← Back to archive
          </Link>
          <div className="mt-5 flex flex-wrap items-center gap-2">
            <span className={`rounded-full px-3 py-1 text-xs font-medium uppercase tracking-[0.18em] ${zoneTone.badge}`}>
              {zoneTone.label}
            </span>
            <span className="rounded-full border border-border/70 px-3 py-1 text-xs uppercase tracking-[0.18em] text-muted-foreground">
              {paper.discipline}
            </span>
            <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
              {formatUnixDate(paper.created_at)}
            </span>
          </div>

          <h1 className="mt-5 font-serif text-4xl tracking-tight text-foreground sm:text-5xl">
            {paper.title}
          </h1>
          <p className="mt-3 text-sm text-muted-foreground">
            {paper.author_name} · DOI {paper.doi || "pending"} · promoted{" "}
            {formatUnixDate(paper.promoted_at)}
          </p>

          {splitKeywords(paper.keywords).length > 0 ? (
            <div className="mt-5 flex flex-wrap gap-2">
              {splitKeywords(paper.keywords).map((keyword) => (
                <span
                  key={keyword}
                  className="rounded-full bg-background/75 px-3 py-1 text-xs text-muted-foreground"
                >
                  {keyword}
                </span>
              ))}
            </div>
          ) : null}

          <section className="mt-8 rounded-[1.6rem] border border-border/70 bg-background/65 p-5">
            <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
              Abstract
            </p>
            <p className="mt-3 text-base leading-8 text-foreground">
              {paper.abstract || "No abstract available."}
            </p>
          </section>

          <section className="mt-8 space-y-5">
            <div>
              <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                Full text
              </p>
              {contentBlocks.length > 0 ? (
                <div className="mt-4 space-y-5 text-base leading-8 text-foreground">
                  {contentBlocks.map((block, index) => (
                    <p key={`${paper.id}-${index}`}>{block}</p>
                  ))}
                </div>
              ) : (
                <p className="mt-4 rounded-[1.4rem] border border-dashed border-border/70 bg-background/65 p-5 text-sm leading-7 text-muted-foreground">
                  This paper does not expose body content yet. The detail page
                  still preserves the governance rail and metadata so the route
                  remains usable while content ingestion catches up.
                </p>
              )}
            </div>
          </section>
        </article>

        <aside className="space-y-4">
          <section className="rounded-[1.8rem] border border-border/80 bg-card/75 p-5">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Governance Rail
            </p>
            <dl className="mt-4 grid gap-3">
              <div className="rounded-[1.2rem] border border-border/70 bg-background/70 p-4">
                <dt className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Score</dt>
                <dd className="mt-2 text-2xl font-semibold text-foreground">{formatScore(paper.shit_score)}</dd>
              </div>
              <div className="rounded-[1.2rem] border border-border/70 bg-background/70 p-4">
                <dt className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Average rating</dt>
                <dd className="mt-2 text-2xl font-semibold text-foreground">
                  {formatScore(paper.avg_rating)} / {paper.rating_count}
                </dd>
              </div>
              <div className="rounded-[1.2rem] border border-border/70 bg-background/70 p-4">
                <dt className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Views · controversy</dt>
                <dd className="mt-2 text-sm text-foreground">
                  {formatCompactNumber(paper.view_count)} views · {formatScore(paper.controversy_index)} controversy
                </dd>
              </div>
            </dl>
          </section>

          <RatingComposer paperId={paper.id} />
          <FlagPaperForm paperId={paper.id} />

          <section className="rounded-[1.8rem] border border-border/80 bg-card/75 p-5">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
              Ratings Feed
            </p>
            {!ratingsResult.ok ? (
              <p className="mt-4 text-sm leading-6 text-muted-foreground">
                {ratingsResult.error.detail}
              </p>
            ) : ratingsResult.data.items.length === 0 ? (
              <p className="mt-4 text-sm leading-6 text-muted-foreground">
                No ratings yet. This is the empty state the detail route keeps
                stable until the first governance action lands.
              </p>
            ) : (
              <ul className="mt-4 space-y-3">
                {ratingsResult.data.items.map((rating) => (
                  <li
                    key={rating.id}
                    className="rounded-[1.2rem] border border-border/70 bg-background/70 p-4"
                  >
                    <p className="text-sm font-semibold text-foreground">
                      {rating.nickname || rating.username} · {rating.score}/10
                    </p>
                    <p className="mt-1 text-xs text-muted-foreground">
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
        </aside>
      </Container>
    </div>
  );
}
