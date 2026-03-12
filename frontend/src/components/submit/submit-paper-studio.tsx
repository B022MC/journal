"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { apiFetch } from "@/lib/api";
import { readBrowserToken } from "@/lib/auth/browser-session";
import type {
  SubmitPaperRequest,
  SubmitPaperResponse,
} from "@/lib/journal/contracts";
import { splitKeywords } from "@/lib/journal/presenters";

const disciplines = [
  { value: "science", label: "Science" },
  { value: "humanities", label: "Humanities" },
  { value: "information", label: "Information" },
  { value: "technology", label: "Technology" },
  { value: "other", label: "Other" },
];

type SubmitDraft = Required<SubmitPaperRequest>;
type FieldErrorMap = Partial<Record<keyof SubmitDraft, string>>;

const emptyDraft: SubmitDraft = {
  title: "",
  title_en: "",
  abstract: "",
  abstract_en: "",
  content: "",
  discipline: "",
  keywords: "",
};

const inputClassName =
  "mt-2 w-full rounded-[1.25rem] border border-border/80 bg-card px-3 py-3 text-foreground";

function validateDraft(draft: SubmitDraft) {
  const fieldErrors: FieldErrorMap = {};

  if (draft.title.trim().length < 6) {
    fieldErrors.title = "Use at least 6 characters so the archive card has a stable title.";
  }
  if (!draft.discipline.trim()) {
    fieldErrors.discipline = "Choose a discipline before submission.";
  }
  if (draft.abstract.trim().length < 24) {
    fieldErrors.abstract = "The abstract should explain the claim in at least 24 characters.";
  }
  if (draft.content.trim().length < 120) {
    fieldErrors.content = "The body should contain at least 120 characters so the preview is meaningful.";
  }

  return fieldErrors;
}

function previewParagraphs(content: string) {
  return content
    .split(/\n{2,}/)
    .map((block) => block.trim())
    .filter(Boolean);
}

export function SubmitPaperStudio({
  authorName,
}: {
  authorName: string;
}) {
  const router = useRouter();
  const [activePanel, setActivePanel] = useState<"edit" | "preview">("edit");
  const [draft, setDraft] = useState<SubmitDraft>(emptyDraft);
  const [fieldErrors, setFieldErrors] = useState<FieldErrorMap>({});
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  const keywords = splitKeywords(draft.keywords);
  const paragraphs = previewParagraphs(draft.content);

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between gap-3 rounded-[1.6rem] border border-border/80 bg-card/80 p-4 lg:hidden">
        <div>
          <p className="text-sm font-semibold text-foreground">Mobile layout</p>
          <p className="text-xs leading-6 text-muted-foreground">
            Switch between editor and preview without losing the draft.
          </p>
        </div>
        <div className="inline-flex rounded-full border border-border/80 bg-background/80 p-1">
          <button
            type="button"
            className={`rounded-full px-3 py-1 text-xs font-medium uppercase tracking-[0.18em] ${activePanel === "edit" ? "bg-primary text-primary-foreground" : "text-muted-foreground"}`}
            onClick={() => setActivePanel("edit")}
          >
            Edit
          </button>
          <button
            type="button"
            className={`rounded-full px-3 py-1 text-xs font-medium uppercase tracking-[0.18em] ${activePanel === "preview" ? "bg-primary text-primary-foreground" : "text-muted-foreground"}`}
            onClick={() => setActivePanel("preview")}
          >
            Preview
          </button>
        </div>
      </div>

      <div className="grid gap-5 lg:grid-cols-[minmax(0,0.95fr)_minmax(320px,0.75fr)]">
        <form
          className={`${activePanel === "preview" ? "hidden lg:block" : "block"} rounded-[2rem] border border-border/80 bg-card/90 p-6 sm:p-8`}
          onSubmit={(event) => {
            event.preventDefault();
            const nextErrors = validateDraft(draft);

            setFieldErrors(nextErrors);
            setError(null);
            setMessage(null);

            if (Object.keys(nextErrors).length > 0) {
              setError("Fix the highlighted fields. Validation failures keep the current draft in place.");
              return;
            }

            const token = readBrowserToken();
            if (!token) {
              router.push("/login?reason=protected&returnTo=/submit");
              router.refresh();
              return;
            }

            startTransition(async () => {
              const result = await apiFetch<SubmitPaperResponse>("/papers/submit", {
                access: "required",
                method: "POST",
                token,
                body: {
                  title: draft.title.trim(),
                  title_en: draft.title_en.trim(),
                  abstract: draft.abstract.trim(),
                  abstract_en: draft.abstract_en.trim(),
                  content: draft.content.trim(),
                  discipline: draft.discipline,
                  keywords: draft.keywords.trim(),
                },
              });

              if (!result.ok) {
                setError(result.error.detail);
                return;
              }

              setMessage(
                `Paper queued as #${result.data.id}${result.data.doi ? ` with DOI ${result.data.doi}` : ""}. Redirecting to your workspace.`,
              );
              router.push(`/me?submitted=${result.data.id}`);
              router.refresh();
            });
          }}
        >
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
                Submit
              </p>
              <h2 className="mt-3 font-serif text-3xl tracking-tight text-foreground">
                Draft in one place, preview in another.
              </h2>
            </div>
            <div className="rounded-[1.2rem] border border-border/70 bg-background/70 px-4 py-3 text-sm text-muted-foreground">
              Signed in as {authorName}
            </div>
          </div>

          <div className="mt-8 grid gap-5">
            <label className="block text-sm text-muted-foreground">
              Title
              <input
                required
                value={draft.title}
                className={inputClassName}
                onChange={(event) =>
                  setDraft((current) => ({ ...current, title: event.target.value }))
                }
                placeholder="A deliberately provocative but still searchable title"
              />
              {fieldErrors.title ? (
                <p className="mt-2 text-xs text-[#8b312e]">{fieldErrors.title}</p>
              ) : null}
            </label>

            <label className="block text-sm text-muted-foreground">
              English title
              <input
                value={draft.title_en}
                className={inputClassName}
                onChange={(event) =>
                  setDraft((current) => ({ ...current, title_en: event.target.value }))
                }
                placeholder="Optional translation"
              />
            </label>

            <div className="grid gap-5 md:grid-cols-[minmax(0,1fr)_220px]">
              <label className="block text-sm text-muted-foreground">
                Discipline
                <select
                  required
                  value={draft.discipline}
                  className={inputClassName}
                  onChange={(event) =>
                    setDraft((current) => ({
                      ...current,
                      discipline: event.target.value,
                    }))
                  }
                >
                  <option value="">Choose a discipline</option>
                  {disciplines.map((discipline) => (
                    <option key={discipline.value} value={discipline.value}>
                      {discipline.label}
                    </option>
                  ))}
                </select>
                {fieldErrors.discipline ? (
                  <p className="mt-2 text-xs text-[#8b312e]">
                    {fieldErrors.discipline}
                  </p>
                ) : null}
              </label>

              <label className="block text-sm text-muted-foreground">
                Keywords
                <input
                  value={draft.keywords}
                  className={inputClassName}
                  onChange={(event) =>
                    setDraft((current) => ({ ...current, keywords: event.target.value }))
                  }
                  placeholder="search, governance, archive"
                />
              </label>
            </div>

            <label className="block text-sm text-muted-foreground">
              Abstract
              <textarea
                required
                rows={5}
                value={draft.abstract}
                className={inputClassName}
                onChange={(event) =>
                  setDraft((current) => ({ ...current, abstract: event.target.value }))
                }
                placeholder="Summarize the claim, why it matters, and what makes it contentious."
              />
              {fieldErrors.abstract ? (
                <p className="mt-2 text-xs text-[#8b312e]">
                  {fieldErrors.abstract}
                </p>
              ) : null}
            </label>

            <label className="block text-sm text-muted-foreground">
              English abstract
              <textarea
                rows={4}
                value={draft.abstract_en}
                className={inputClassName}
                onChange={(event) =>
                  setDraft((current) => ({
                    ...current,
                    abstract_en: event.target.value,
                  }))
                }
                placeholder="Optional translation"
              />
            </label>

            <label className="block text-sm text-muted-foreground">
              Full text
              <textarea
                required
                rows={12}
                value={draft.content}
                className={inputClassName}
                onChange={(event) =>
                  setDraft((current) => ({ ...current, content: event.target.value }))
                }
                placeholder="Paste the body, methods, observations, and caveats."
              />
              {fieldErrors.content ? (
                <p className="mt-2 text-xs text-[#8b312e]">{fieldErrors.content}</p>
              ) : null}
            </label>
          </div>

          {error ? <p className="mt-5 text-sm text-[#8b312e]">{error}</p> : null}
          {message ? <p className="mt-5 text-sm text-[#426b54]">{message}</p> : null}

          <div className="mt-6 flex flex-wrap items-center gap-3">
            <button
              type="submit"
              disabled={isPending}
              className="inline-flex rounded-full bg-primary px-5 py-3 text-sm font-medium text-primary-foreground disabled:opacity-60"
            >
              {isPending ? "Submitting…" : "Submit to archive"}
            </button>
            <p className="text-sm text-muted-foreground">
              Failed validation or API errors keep the draft intact.
            </p>
          </div>
        </form>

        <aside
          className={`${activePanel === "edit" ? "hidden lg:block" : "block"} rounded-[2rem] border border-border/80 bg-secondary/75 p-6 sm:p-7`}
        >
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Preview
          </p>
          <h2 className="mt-4 font-serif text-3xl tracking-tight text-foreground">
            {draft.title.trim() || "Title preview"}
          </h2>
          <p className="mt-3 text-sm text-muted-foreground">
            {draft.discipline
              ? disciplines.find((item) => item.value === draft.discipline)?.label
              : "Discipline pending"}
            {" · "}
            {authorName}
          </p>

          {keywords.length > 0 ? (
            <div className="mt-5 flex flex-wrap gap-2">
              {keywords.map((keyword) => (
                <span
                  key={keyword}
                  className="rounded-full bg-background/80 px-3 py-1 text-xs text-muted-foreground"
                >
                  {keyword}
                </span>
              ))}
            </div>
          ) : null}

          <section className="mt-6 rounded-[1.4rem] border border-border/70 bg-background/70 p-4">
            <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
              Abstract
            </p>
            <p className="mt-3 text-sm leading-7 text-foreground">
              {draft.abstract.trim() ||
                "The abstract preview appears here as soon as the draft explains the core claim."}
            </p>
          </section>

          <section className="mt-5 rounded-[1.4rem] border border-border/70 bg-background/70 p-4">
            <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
              Body preview
            </p>
            {paragraphs.length > 0 ? (
              <div className="mt-3 space-y-4 text-sm leading-7 text-foreground">
                {paragraphs.slice(0, 4).map((paragraph, index) => (
                  <p key={`${index}-${paragraph.slice(0, 24)}`}>{paragraph}</p>
                ))}
              </div>
            ) : (
              <p className="mt-3 text-sm leading-7 text-muted-foreground">
                Add at least one paragraph to inspect the reading surface before submission.
              </p>
            )}
          </section>

          <section className="mt-5 rounded-[1.4rem] border border-border/70 bg-background/70 p-4">
            <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
              Submission checks
            </p>
            <ul className="mt-3 space-y-2 text-sm leading-6 text-foreground">
              <li>Required fields block submission before the request leaves the browser.</li>
              <li>Authentication failures send the user back through the shared login route.</li>
              <li>The success path returns to `/me` so the workflow closes on the workspace.</li>
            </ul>
          </section>
        </aside>
      </div>
    </div>
  );
}
