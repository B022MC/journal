import type { ReactNode } from "react";
import Link from "next/link";
import { Container } from "@/components/layout/container";
import { cn } from "@/lib/utils";
import type { ApiRuntimeSnapshot } from "@/lib/api/runtime";

const navItems = [
  { label: "Home", href: "/", state: "live" as const, note: "Baseline shell" },
  { label: "Papers", state: "planned" as const, note: "Phase 1" },
  { label: "Search", state: "planned" as const, note: "Search ADR first" },
  { label: "Login", state: "planned" as const, note: "Phase 1" },
  { label: "Submit", state: "planned" as const, note: "Phase 2" },
];

function HeaderChip({
  children,
  tone = "neutral",
}: {
  children: ReactNode;
  tone?: "neutral" | "warn";
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-3 py-1 text-[11px] font-medium uppercase tracking-[0.18em]",
        tone === "warn"
          ? "border-[#b06a2d]/35 bg-[#b06a2d]/10 text-[#8a4b2a]"
          : "border-border/80 bg-background/70 text-muted-foreground",
      )}
    >
      {children}
    </span>
  );
}

export function SiteHeader({ runtime }: { runtime: ApiRuntimeSnapshot }) {
  return (
    <header className="sticky top-0 z-30 border-b border-border/70 bg-background/90 backdrop-blur-xl">
      <Container className="flex flex-col gap-4 py-4 lg:flex-row lg:items-center lg:justify-between">
        <div className="space-y-2">
          <Link
            href="/"
            className="inline-flex items-center gap-3 text-foreground transition-colors hover:text-[#8a4b2a]"
          >
            <span className="font-serif text-2xl font-semibold tracking-tight sm:text-3xl">
              S.H.I.T Journal
            </span>
            <HeaderChip tone="warn">Archive Lab</HeaderChip>
          </Link>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">
            Main-site shell for reading, search, and governance transparency.
            The roadmap splash page no longer defines the delivery baseline.
          </p>
        </div>

        <div className="flex flex-col gap-3 lg:items-end">
          <nav className="flex flex-wrap gap-2">
            {navItems.map((item) =>
              item.state === "live" ? (
                <Link
                  key={item.label}
                  href={item.href}
                  className="inline-flex items-center gap-2 rounded-full border border-border/80 bg-card/70 px-3 py-2 text-sm text-foreground transition-colors hover:border-[#8a4b2a]/40 hover:text-[#8a4b2a]"
                >
                  {item.label}
                  <span className="text-[10px] uppercase tracking-[0.18em] text-muted-foreground">
                    {item.note}
                  </span>
                </Link>
              ) : (
                <span
                  key={item.label}
                  className="inline-flex items-center gap-2 rounded-full border border-border/70 bg-background/60 px-3 py-2 text-sm text-muted-foreground"
                >
                  {item.label}
                  <span className="text-[10px] uppercase tracking-[0.18em]">
                    {item.note}
                  </span>
                </span>
              ),
            )}
          </nav>

          <div className="flex flex-wrap gap-2">
            <HeaderChip tone={runtime.source === "env" ? "neutral" : "warn"}>
              {runtime.source === "env"
                ? `API ${runtime.origin}`
                : "API same-origin fallback"}
            </HeaderChip>
            <HeaderChip>
              {runtime.authStrategy === "cookie-bridge"
                ? "Cookie bridge reserved"
                : "Browser token compatibility"}
            </HeaderChip>
          </div>
        </div>
      </Container>
    </header>
  );
}
