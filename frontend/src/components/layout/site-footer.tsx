import { Container } from "@/components/layout/container";
import type { ApiRuntimeSnapshot } from "@/lib/api/runtime";

export function SiteFooter({ runtime }: { runtime: ApiRuntimeSnapshot }) {
  return (
    <footer className="border-t border-border/70 bg-card/60">
      <Container className="grid gap-6 py-8 lg:grid-cols-[minmax(0,1fr)_360px]">
        <section>
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Phase 0 Baseline
          </p>
          <h2 className="mt-3 font-serif text-2xl tracking-tight text-foreground">
            Shell, runtime contract, and failure surfaces are now shared
            infrastructure.
          </h2>
          <p className="mt-3 max-w-2xl text-sm leading-7 text-muted-foreground">
            Phase 1 pages inherit this frame. Search work stays behind its ADR
            and FULLTEXT fallback until quality gates are met. Scope expansion
            now requires both design and execution snapshot updates.
          </p>
        </section>

        <aside className="rounded-[1.6rem] border border-border/80 bg-background/70 p-5">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Diagnostics
          </p>
          <ul className="mt-4 space-y-3 text-sm leading-6 text-foreground">
            {runtime.diagnostics.map((item) => (
              <li
                key={item}
                className="rounded-[1rem] border border-border/70 bg-card/70 px-3 py-3"
              >
                {item}
              </li>
            ))}
          </ul>
        </aside>
      </Container>
    </footer>
  );
}
