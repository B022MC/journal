import Link from "next/link";
import type { PaperItem } from "@/lib/journal/contracts";
import {
  excerpt,
  formatCompactNumber,
  formatScore,
  formatUnixDate,
  getZoneTone,
  splitKeywords,
} from "@/lib/journal/presenters";

export function PaperCard({
  paper,
  compact = false,
}: {
  paper: PaperItem;
  compact?: boolean;
}) {
  const zoneTone = getZoneTone(paper.zone);
  const keywords = splitKeywords(paper.keywords);

  return (
    <article
      className={`rounded-[1.8rem] border ${zoneTone.border} bg-card/85 p-5 shadow-[0_22px_60px_rgba(23,20,17,0.06)]`}
    >
      <div className="flex flex-wrap items-center gap-2">
        <span className={`rounded-full px-3 py-1 text-xs font-medium uppercase tracking-[0.18em] ${zoneTone.badge}`}>
          {zoneTone.label}
        </span>
        <span className="rounded-full border border-border/70 px-3 py-1 text-xs uppercase tracking-[0.18em] text-muted-foreground">
          {paper.discipline || "Undisciplined"}
        </span>
        <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
          {formatUnixDate(paper.created_at)}
        </span>
      </div>

      <div className="mt-4 space-y-3">
        <Link href={`/papers/${paper.id}`} className="group block">
          <h2 className="font-serif text-2xl tracking-tight text-foreground transition-colors group-hover:text-[#8a4b2a]">
            {paper.title}
          </h2>
        </Link>
        <p className="text-sm text-muted-foreground">
          {paper.author_name} · DOI {paper.doi || "pending"}
        </p>
        <p className="text-sm leading-7 text-muted-foreground">
          {excerpt(paper.abstract || paper.abstract_en || "", compact ? 140 : 200)}
        </p>
      </div>

      {keywords.length > 0 ? (
        <div className="mt-4 flex flex-wrap gap-2">
          {keywords.map((keyword) => (
            <span
              key={keyword}
              className="rounded-full bg-background/80 px-3 py-1 text-xs text-muted-foreground"
            >
              {keyword}
            </span>
          ))}
        </div>
      ) : null}

      <dl className="mt-5 grid gap-3 border-t border-border/70 pt-4 sm:grid-cols-3">
        <div>
          <dt className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
            S.H.I.T Score
          </dt>
          <dd className="mt-1 text-lg font-semibold text-foreground">
            {formatScore(paper.shit_score)}
          </dd>
        </div>
        <div>
          <dt className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
            Ratings
          </dt>
          <dd className="mt-1 text-lg font-semibold text-foreground">
            {formatScore(paper.avg_rating)} / {paper.rating_count}
          </dd>
        </div>
        <div>
          <dt className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
            Views
          </dt>
          <dd className="mt-1 text-lg font-semibold text-foreground">
            {formatCompactNumber(paper.view_count)}
          </dd>
        </div>
      </dl>
    </article>
  );
}
