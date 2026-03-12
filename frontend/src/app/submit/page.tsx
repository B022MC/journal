import Link from "next/link";
import { Container } from "@/components/layout/container";
import { PageErrorState } from "@/components/states/page-error-state";
import { SubmitPaperStudio } from "@/components/submit/submit-paper-studio";
import { requireCurrentUser } from "@/lib/journal/protected";

export const dynamic = "force-dynamic";

export default async function SubmitPage() {
  const currentUser = await requireCurrentUser("/submit");
  if (!currentUser.ok) {
    return (
      <div className="py-10 sm:py-12">
        <Container>
          <PageErrorState
            title="The submit workspace could not be opened."
            detail={currentUser.error.detail}
            actionHref="/papers"
            actionLabel="Back to archive"
          />
        </Container>
      </div>
    );
  }

  const authorName = currentUser.user.nickname || currentUser.user.username;

  return (
    <div className="py-10 sm:py-12">
      <Container className="space-y-6">
        <section className="rounded-[2rem] border border-border/80 bg-card/90 p-6 shadow-[0_30px_80px_rgba(23,20,17,0.08)] sm:p-8">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Submission Studio
          </p>
          <h1 className="mt-4 max-w-3xl font-serif text-4xl tracking-tight text-foreground sm:text-5xl">
            Publish without leaving the server-first shell behind.
          </h1>
          <p className="mt-4 max-w-3xl text-sm leading-7 text-muted-foreground sm:text-base">
            `/submit` is protected, but the rest of the site stays public. The
            draft editor keeps validation and preview on the client while the
            route itself remains predictable on refresh through the cookie
            bridge.
          </p>
          <div className="mt-6 flex flex-wrap gap-3">
            <Link
              href="/me"
              className="inline-flex rounded-full border border-border/80 bg-background/75 px-5 py-3 text-sm font-medium text-foreground"
            >
              Open workspace
            </Link>
            <Link
              href="/papers"
              className="inline-flex rounded-full border border-border/80 bg-background/75 px-5 py-3 text-sm font-medium text-foreground"
            >
              Back to archive
            </Link>
          </div>
        </section>

        <SubmitPaperStudio authorName={authorName} />
      </Container>
    </div>
  );
}
