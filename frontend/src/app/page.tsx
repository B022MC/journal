import { Container } from "@/components/layout/container";
import { getApiRuntimeSnapshot } from "@/lib/api/runtime";

export default function Home() {
  const runtime = getApiRuntimeSnapshot();
  const workstreams = [
    {
      title: "Phase 0 shell",
      detail:
        "New tokens, reusable header/footer/container, and runtime diagnostics are now the default frame for every page.",
      status: "Active now",
    },
    {
      title: "Phase 1 routes",
      detail:
        "Home, papers, detail, login, and register follow next. They inherit this shell instead of the old roadmap page.",
      status: "Next",
    },
    {
      title: "Search baseline",
      detail:
        "Search ADR and staged engine work stay inside the current execution boundary, with MySQL FULLTEXT still serving as fallback.",
      status: "Pinned",
    },
  ];

  const guardrails = [
    "Server Components remain the default data path. Client state is reserved for interaction islands only.",
    "Authentication keeps a migration slot for an httpOnly cookie bridge instead of forcing the whole site back to CSR.",
    "Missing API origin falls back to same-origin /api/v1, and request failures are normalized by the shared API layer plus global error boundary.",
  ];

  return (
    <div className="pb-16 pt-8 sm:pb-20 sm:pt-10">
      <Container className="grid gap-6 lg:grid-cols-[minmax(0,1.25fr)_minmax(300px,0.75fr)]">
        <section className="overflow-hidden rounded-[2rem] border border-border/80 bg-card/90 p-6 shadow-[0_30px_80px_rgba(23,20,17,0.10)] sm:p-8">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Phase 0 Foundation
          </p>
          <div className="mt-5 max-w-3xl space-y-5">
            <h1 className="max-w-2xl font-serif text-4xl leading-none tracking-tight text-foreground sm:text-5xl lg:text-6xl">
              Archive Lab shell is now the baseline, not the roadmap splash page.
            </h1>
            <p className="max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg">
              This surface intentionally stops at the platform layer. It locks
              typography, color, spacing, header/footer reuse, and the API
              runtime contract before the Phase 1 reading flow lands.
            </p>
          </div>
          <div className="mt-8 grid gap-4 md:grid-cols-3">
            {workstreams.map((item) => (
              <article
                key={item.title}
                className="rounded-[1.4rem] border border-border/80 bg-background/75 p-4"
              >
                <p className="text-xs font-medium uppercase tracking-[0.24em] text-muted-foreground">
                  {item.status}
                </p>
                <h2 className="mt-3 text-xl font-semibold tracking-tight text-foreground">
                  {item.title}
                </h2>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">
                  {item.detail}
                </p>
              </article>
            ))}
          </div>
        </section>

        <aside className="rounded-[2rem] border border-border/80 bg-secondary/70 p-6 shadow-[0_24px_60px_rgba(23,20,17,0.08)] sm:p-7">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Runtime Snapshot
          </p>
          <dl className="mt-5 space-y-4 text-sm">
            <div className="rounded-[1.2rem] border border-border/70 bg-background/75 p-4">
              <dt className="text-muted-foreground">API base</dt>
              <dd className="mt-1 font-mono text-xs text-foreground sm:text-sm">
                {runtime.baseUrl}
              </dd>
            </div>
            <div className="rounded-[1.2rem] border border-border/70 bg-background/75 p-4">
              <dt className="text-muted-foreground">Auth strategy</dt>
              <dd className="mt-1 text-foreground">
                {runtime.authStrategy === "cookie-bridge"
                  ? "Cookie bridge reserved"
                  : "Browser token compatibility mode"}
              </dd>
            </div>
            <div className="rounded-[1.2rem] border border-border/70 bg-background/75 p-4">
              <dt className="text-muted-foreground">Failure handling</dt>
              <dd className="mt-1 leading-6 text-foreground">
                Shared API requests normalize config, auth, network, and HTTP
                failures into stable diagnostics. Unhandled render failures fall
                through the global error boundary.
              </dd>
            </div>
          </dl>
        </aside>
      </Container>

      <Container className="mt-6 grid gap-6 lg:grid-cols-[minmax(0,1fr)_340px]">
        <section className="rounded-[2rem] border border-border/80 bg-card/85 p-6 sm:p-8">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Delivery Guardrails
          </p>
          <ul className="mt-5 space-y-4">
            {guardrails.map((item) => (
              <li
                key={item}
                className="rounded-[1.25rem] border border-border/70 bg-background/70 px-4 py-4 text-sm leading-6 text-foreground"
              >
                {item}
              </li>
            ))}
          </ul>
        </section>

        <aside className="rounded-[2rem] border border-border/80 bg-card/70 p-6">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Expected Env
          </p>
          <ul className="mt-5 space-y-3 text-sm text-muted-foreground">
            <li>
              <span className="font-mono text-foreground">
                JOURNAL_API_ORIGIN
              </span>
              {" "}
              or
              {" "}
              <span className="font-mono text-foreground">
                NEXT_PUBLIC_JOURNAL_API_ORIGIN
              </span>
            </li>
            <li>
              <span className="font-mono text-foreground">
                JOURNAL_API_PREFIX
              </span>
              {" "}
              for non-standard gateways. Default stays at
              {" "}
              <span className="font-mono text-foreground">/api/v1</span>.
            </li>
            <li>
              <span className="font-mono text-foreground">
                JOURNAL_AUTH_STRATEGY
              </span>
              {" "}
              controls migration intent:
              {" "}
              <span className="font-mono text-foreground">cookie-bridge</span>
              {" "}
              or
              {" "}
              <span className="font-mono text-foreground">browser-token</span>.
            </li>
          </ul>
        </aside>
      </Container>
    </div>
  );
}
