import Link from "next/link";

export function PageErrorState({
  detail,
  title = "The page could not load its live data.",
  actionHref,
  actionLabel,
}: {
  detail: string;
  title?: string;
  actionHref?: string;
  actionLabel?: string;
}) {
  return (
    <section className="rounded-[1.8rem] border border-[#8b312e]/20 bg-[#8b312e]/6 p-6 sm:p-8">
      <p className="text-xs font-medium uppercase tracking-[0.28em] text-[#8b312e]">
        Runtime Error
      </p>
      <h2 className="mt-4 font-serif text-3xl tracking-tight text-foreground">
        {title}
      </h2>
      <p className="mt-3 max-w-2xl text-sm leading-7 text-muted-foreground sm:text-base">
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
