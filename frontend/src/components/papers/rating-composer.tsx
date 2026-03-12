"use client";

import Link from "next/link";
import { useState, useTransition } from "react";
import { apiFetch } from "@/lib/api";
import { readBrowserToken } from "@/lib/auth/browser-session";
import type { CommonResponse } from "@/lib/journal/contracts";

export function RatingComposer({
  paperId,
}: {
  paperId: number;
}) {
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  return (
    <form
      className="space-y-3 rounded-[1.4rem] border border-border/70 bg-background/70 p-4"
      onSubmit={(event) => {
        event.preventDefault();
        const form = event.currentTarget;
        const formData = new FormData(form);
        const token = readBrowserToken();

        setError(null);
        setMessage(null);

        if (!token) {
          setError("Sign in first. Rating uses the current browser-token compatibility path.");
          return;
        }

        startTransition(async () => {
          const result = await apiFetch<CommonResponse>(`/papers/${paperId}/rate`, {
            access: "required",
            body: {
              score: Number(formData.get("score")),
              comment: String(formData.get("comment") ?? ""),
            },
            method: "POST",
            token,
          });

          if (!result.ok) {
            setError(result.error.detail);
            return;
          }

          setMessage(result.data.message || "Your rating was submitted.");
          form.reset();
        });
      }}
    >
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-sm font-semibold text-foreground">Rate this paper</p>
          <p className="text-xs text-muted-foreground">
            This interaction stays client-side while the page itself remains server-rendered.
          </p>
        </div>
        <Link
          href={`/login?returnTo=/papers/${paperId}`}
          className="text-xs uppercase tracking-[0.18em] text-[#8a4b2a]"
        >
          Sign in
        </Link>
      </div>

      <label className="block text-sm text-muted-foreground">
        Score
        <select
          name="score"
          defaultValue="7"
          className="mt-2 w-full rounded-2xl border border-border/80 bg-card px-3 py-2 text-foreground"
        >
          {Array.from({ length: 10 }).map((_, index) => {
            const score = index + 1;
            return (
              <option key={score} value={score}>
                {score}
              </option>
            );
          })}
        </select>
      </label>

      <label className="block text-sm text-muted-foreground">
        Comment
        <textarea
          name="comment"
          rows={4}
          className="mt-2 w-full rounded-[1.25rem] border border-border/80 bg-card px-3 py-3 text-foreground"
          placeholder="What changed your reading of this paper?"
        />
      </label>

      {error ? <p className="text-sm text-[#8b312e]">{error}</p> : null}
      {message ? <p className="text-sm text-[#426b54]">{message}</p> : null}

      <button
        type="submit"
        disabled={isPending}
        className="inline-flex rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground disabled:opacity-60"
      >
        {isPending ? "Submitting…" : "Submit rating"}
      </button>
    </form>
  );
}
