import type { ReactNode } from "react";
import Link from "next/link";
import { LogoutButton } from "@/components/auth/logout-button";
import { Container } from "@/components/layout/container";
import { cn } from "@/lib/utils";
import type { ApiRuntimeSnapshot } from "@/lib/api/runtime";
import type { UserInfo } from "@/lib/journal/contracts";

const primaryNavItems = [
  { label: "Home", href: "/", note: "Live" },
  { label: "Papers", href: "/papers", note: "Archive" },
  { label: "Submit", href: "/submit", note: "Protected" },
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

export function SiteHeader({
  runtime,
  currentUser,
}: {
  runtime: ApiRuntimeSnapshot;
  currentUser: UserInfo | null;
}) {
  const displayName = currentUser?.nickname || currentUser?.username || null;

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
            Main-site shell for reading, search, and governance transparency,
            with release flags that can swap the homepage between the live desk
            and the roadmap rollback surface.
          </p>
        </div>

        <div className="flex flex-col gap-3 lg:items-end">
          <nav className="flex flex-wrap gap-2">
            {primaryNavItems.map((item) => (
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
            ))}
            {currentUser ? (
              <Link
                href="/me"
                className="inline-flex items-center gap-2 rounded-full border border-border/80 bg-card/70 px-3 py-2 text-sm text-foreground transition-colors hover:border-[#8a4b2a]/40 hover:text-[#8a4b2a]"
              >
                Workspace
                <span className="text-[10px] uppercase tracking-[0.18em] text-muted-foreground">
                  Live
                </span>
              </Link>
            ) : (
              <>
                <Link
                  href="/login?returnTo=/me"
                  className="inline-flex items-center gap-2 rounded-full border border-border/80 bg-card/70 px-3 py-2 text-sm text-foreground transition-colors hover:border-[#8a4b2a]/40 hover:text-[#8a4b2a]"
                >
                  Login
                  <span className="text-[10px] uppercase tracking-[0.18em] text-muted-foreground">
                    Auth
                  </span>
                </Link>
                <Link
                  href="/register?returnTo=/submit"
                  className="inline-flex items-center gap-2 rounded-full border border-border/80 bg-card/70 px-3 py-2 text-sm text-foreground transition-colors hover:border-[#8a4b2a]/40 hover:text-[#8a4b2a]"
                >
                  Register
                  <span className="text-[10px] uppercase tracking-[0.18em] text-muted-foreground">
                    Access
                  </span>
                </Link>
              </>
            )}
          </nav>

          <div className="flex flex-wrap items-center justify-end gap-2">
            <HeaderChip tone={runtime.source === "env" ? "neutral" : "warn"}>
              {runtime.source === "env"
                ? `API ${runtime.origin}`
                : "API same-origin fallback"}
            </HeaderChip>
            <HeaderChip>
              {runtime.authStrategy === "cookie-bridge"
                ? "Cookie bridge active"
                : "Browser token compatibility"}
            </HeaderChip>
            <HeaderChip tone={currentUser ? "neutral" : "warn"}>
              {displayName ? `Signed in as ${displayName}` : "Anonymous reader"}
            </HeaderChip>
            {currentUser ? (
              <Link
                href={`/users/${currentUser.id}`}
                className="inline-flex rounded-full border border-border/80 bg-background/80 px-3 py-1.5 text-xs font-medium uppercase tracking-[0.18em] text-foreground transition-colors hover:border-[#8a4b2a]/40 hover:text-[#8a4b2a]"
              >
                Public profile
              </Link>
            ) : null}
            {currentUser ? <LogoutButton className="px-3 py-1.5 text-xs uppercase tracking-[0.18em]" /> : null}
          </div>
        </div>
      </Container>
    </header>
  );
}
