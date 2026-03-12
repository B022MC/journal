"use client";

import { Button } from "@/components/ui/button";
import { toApiDiagnostic } from "@/lib/api/errors";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  const diagnostic = toApiDiagnostic(error);

  return (
    <html lang="en">
      <body className="min-h-screen bg-[#f3efe6] text-[#171411] antialiased">
        <main className="mx-auto flex min-h-screen w-full max-w-3xl items-center px-4 py-10 sm:px-6">
          <section className="w-full rounded-[2rem] border border-[rgba(23,20,17,0.14)] bg-[rgba(250,246,239,0.92)] p-6 shadow-[0_24px_60px_rgba(23,20,17,0.08)] sm:p-8">
            <p className="text-xs font-medium uppercase tracking-[0.28em] text-[#6b6258]">
              Global Render Failure
            </p>
            <h1 className="mt-4 text-4xl font-semibold tracking-tight">
              {diagnostic.title}
            </h1>
            <p className="mt-4 text-sm leading-7 text-[#6b6258] sm:text-base">
              {diagnostic.detail}
            </p>
            {diagnostic.digest ? (
              <p className="mt-4 font-mono text-xs text-[#6b6258]">
                digest: {diagnostic.digest}
              </p>
            ) : null}
            <div className="mt-8">
              <Button onClick={() => reset()}>Retry application shell</Button>
            </div>
          </section>
        </main>
      </body>
    </html>
  );
}
