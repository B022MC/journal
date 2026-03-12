import Link from "next/link";
import { Container } from "@/components/layout/container";
import { AuthForm } from "@/components/auth/auth-form";
import { parseSearchParam } from "@/lib/journal/presenters";

export default async function RegisterPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const params = await searchParams;
  const returnTo = parseSearchParam(params.returnTo, "/");

  return (
    <div className="py-10 sm:py-12">
      <Container className="grid gap-6 lg:grid-cols-[minmax(0,0.9fr)_minmax(320px,0.7fr)]">
        <section className="rounded-[2rem] border border-border/80 bg-card/90 p-6 sm:p-8">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Register
          </p>
          <h1 className="mt-4 font-serif text-4xl tracking-tight text-foreground sm:text-5xl">
            Join the archive before you rate, report, or publish.
          </h1>
          <p className="mt-4 max-w-2xl text-sm leading-7 text-muted-foreground sm:text-base">
            Registration stays explicit about the product: this is an archive
            with governance, not a generic social feed. The form lands you back
            in the login flow with a predictable redirect target.
          </p>

          <div className="mt-8">
            <AuthForm mode="register" returnTo={returnTo} />
          </div>

          <p className="mt-6 text-sm text-muted-foreground">
            Already registered?{" "}
            <Link href={`/login?returnTo=${encodeURIComponent(returnTo)}`} className="text-[#8a4b2a]">
              Sign in instead
            </Link>
          </p>
        </section>

        <aside className="rounded-[2rem] border border-border/80 bg-secondary/75 p-6 sm:p-7">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Orientation
          </p>
          <ul className="mt-5 space-y-4 text-sm leading-7 text-foreground">
            <li>Readers can browse before they sign in, which keeps the archive index public.</li>
            <li>Scoring and flagging both explain their consequences before the action is attempted.</li>
            <li>Once Phase 2 lands, this account flows directly into submit and personal workspace routes.</li>
          </ul>
        </aside>
      </Container>
    </div>
  );
}
