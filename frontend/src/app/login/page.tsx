import Link from "next/link";
import { Container } from "@/components/layout/container";
import { AuthForm } from "@/components/auth/auth-form";
import { parseSearchParam } from "@/lib/journal/presenters";

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const params = await searchParams;
  const returnTo = parseSearchParam(params.returnTo, "/");
  const registered = parseSearchParam(params.registered) === "1";

  return (
    <div className="py-10 sm:py-12">
      <Container className="grid gap-6 lg:grid-cols-[minmax(0,0.9fr)_minmax(320px,0.7fr)]">
        <section className="rounded-[2rem] border border-border/80 bg-card/90 p-6 sm:p-8">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Login
          </p>
          <h1 className="mt-4 font-serif text-4xl tracking-tight text-foreground sm:text-5xl">
            Re-enter the archive without turning the whole site into a client app.
          </h1>
          <p className="mt-4 max-w-2xl text-sm leading-7 text-muted-foreground sm:text-base">
            Sign-in stays on its own route. Browse and paper detail remain
            server-rendered; only the form and post-login storage live on the
            client compatibility path.
          </p>

          <div className="mt-8">
            <AuthForm mode="login" registered={registered} returnTo={returnTo} />
          </div>

          <p className="mt-6 text-sm text-muted-foreground">
            No account yet?{" "}
            <Link href={`/register?returnTo=${encodeURIComponent(returnTo)}`} className="text-[#8a4b2a]">
              Create one
            </Link>
          </p>
        </section>

        <aside className="rounded-[2rem] border border-border/80 bg-secondary/75 p-6 sm:p-7">
          <p className="text-xs font-medium uppercase tracking-[0.28em] text-muted-foreground">
            Why this page exists
          </p>
          <ul className="mt-5 space-y-4 text-sm leading-7 text-foreground">
            <li>Community scoring is explicit and never hidden behind a dark admin dashboard.</li>
            <li>Zone movement and reporting stay visible from the reading flow onward.</li>
            <li>The current auth bridge stores the bearer token in the browser until the cookie bridge is wired.</li>
          </ul>
        </aside>
      </Container>
    </div>
  );
}
