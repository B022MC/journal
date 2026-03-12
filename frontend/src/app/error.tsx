"use client";

import { useEffect } from "react";
import Link from "next/link";
import { Container } from "@/components/layout/container";
import { Button, buttonVariants } from "@/components/ui/button";
import { toApiDiagnostic } from "@/lib/api/errors";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  const diagnostic = toApiDiagnostic(error);

  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="py-12 sm:py-16">
      <Container>
        <section className="rounded-[2rem] border border-border/80 bg-card/90 p-6 shadow-[0_24px_60px_rgba(23,20,17,0.08)] sm:p-8">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Request or Render Failure
          </p>
          <h1 className="mt-4 font-serif text-4xl tracking-tight text-foreground">
            {diagnostic.title}
          </h1>
          <p className="mt-4 max-w-2xl text-sm leading-7 text-muted-foreground sm:text-base">
            {diagnostic.detail}
          </p>
          {diagnostic.digest ? (
            <p className="mt-4 font-mono text-xs text-muted-foreground">
              digest: {diagnostic.digest}
            </p>
          ) : null}
          <div className="mt-8 flex flex-wrap gap-3">
            <Button onClick={() => reset()}>Retry render</Button>
            <Link href="/" className={buttonVariants({ variant: "outline" })}>
              Back to shell
            </Link>
          </div>
        </section>
      </Container>
    </div>
  );
}
