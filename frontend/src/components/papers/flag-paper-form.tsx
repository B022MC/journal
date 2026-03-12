"use client";

import Link from "next/link";
import { useState, useTransition } from "react";
import { apiFetch } from "@/lib/api";
import { readBrowserToken } from "@/lib/auth/browser-session";
import type { FlagActionResponse } from "@/lib/journal/contracts";

export function FlagPaperForm({
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
          setError("Sign in first. Reporting also uses the current browser-token compatibility path.");
          return;
        }

        startTransition(async () => {
          const result = await apiFetch<FlagActionResponse>(`/papers/${paperId}/flag`, {
            access: "required",
            body: {
              reason: String(formData.get("reason") ?? ""),
              detail: String(formData.get("detail") ?? ""),
            },
            method: "POST",
            token,
          });

          if (!result.ok) {
            setError(result.error.detail);
            return;
          }

          setMessage(result.data.message || "Your report was submitted.");
          form.reset();
        });
      }}
    >
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-sm font-semibold text-foreground">Report governance concern</p>
          <p className="text-xs text-muted-foreground">
            Reports stay in the governance rail so the reading flow stays clean.
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
        Reason
        <select
          name="reason"
          defaultValue="quality_concern"
          className="mt-2 w-full rounded-2xl border border-border/80 bg-card px-3 py-2 text-foreground"
        >
          <option value="quality_concern">Quality concern</option>
          <option value="harmful_claim">Harmful claim</option>
          <option value="spam_or_noise">Spam or noise</option>
          <option value="ethics">Ethics issue</option>
        </select>
      </label>

      <label className="block text-sm text-muted-foreground">
        Detail
        <textarea
          name="detail"
          rows={3}
          className="mt-2 w-full rounded-[1.25rem] border border-border/80 bg-card px-3 py-3 text-foreground"
          placeholder="Optional context for moderators and voters."
        />
      </label>

      {error ? <p className="text-sm text-[#8b312e]">{error}</p> : null}
      {message ? <p className="text-sm text-[#426b54]">{message}</p> : null}

      <button
        type="submit"
        disabled={isPending}
        className="inline-flex rounded-full border border-border/80 bg-background/75 px-4 py-2 text-sm font-medium text-foreground disabled:opacity-60"
      >
        {isPending ? "Sending…" : "Submit report"}
      </button>
    </form>
  );
}
