import { Container } from "@/components/layout/container";

export function PageLoadingSkeleton({
  label,
}: {
  label: string;
}) {
  return (
    <div className="py-10 sm:py-12">
      <Container className="space-y-6">
        <div className="animate-pulse rounded-[2rem] border border-border/70 bg-card/70 p-6 sm:p-8">
          <div className="h-3 w-32 rounded-full bg-secondary" />
          <div className="mt-5 h-10 w-2/3 rounded-full bg-secondary" />
          <div className="mt-4 h-4 w-full rounded-full bg-secondary/80" />
          <div className="mt-3 h-4 w-5/6 rounded-full bg-secondary/70" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 3 }).map((_, index) => (
            <div
              key={`${label}-${index}`}
              className="animate-pulse rounded-[1.6rem] border border-border/70 bg-card/70 p-5"
            >
              <div className="h-3 w-20 rounded-full bg-secondary" />
              <div className="mt-4 h-6 w-4/5 rounded-full bg-secondary/85" />
              <div className="mt-4 h-4 w-full rounded-full bg-secondary/75" />
              <div className="mt-2 h-4 w-2/3 rounded-full bg-secondary/65" />
            </div>
          ))}
        </div>
      </Container>
    </div>
  );
}
