import Link from "next/link";

export function PageEmptyState({
  eyebrow = "Empty State",
  title,
  detail,
  actionHref,
  actionLabel,
}: {
  eyebrow?: string;
  title: string;
  detail: string;
  actionHref?: string;
  actionLabel?: string;
}) {
  return (
    <section className="rounded-[1.8rem] border border-dashed border-border/80 bg-card/70 p-6 text-center sm:p-8">
      <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
        {eyebrow}
      </p>
      <h2 className="mt-4 font-serif text-3xl tracking-tight text-foreground">
        {title}
      </h2>
      <p className="mx-auto mt-3 max-w-2xl text-sm leading-7 text-muted-foreground sm:text-base">
        {detail}
      </p>
      {actionHref && actionLabel ? (
        <Link
          href={actionHref}
          className="mt-6 inline-flex rounded-full border border-border/80 bg-background/75 px-4 py-2 text-sm text-foreground transition-colors hover:border-[#8a4b2a]/40 hover:text-[#8a4b2a]"
        >
          {actionLabel}
        </Link>
      ) : null}
    </section>
  );
}
