import type { AchievementBadge } from "@/lib/journal/contracts";
import { formatUnixDate, getBadgeTierTone } from "@/lib/journal/presenters";

export function AchievementBadges({
  badges,
  emptyTitle = "No badges unlocked yet.",
  emptyDetail = "This account still renders a stable empty state so the profile pages do not collapse when achievements are missing.",
}: {
  badges: AchievementBadge[];
  emptyTitle?: string;
  emptyDetail?: string;
}) {
  if (badges.length === 0) {
    return (
      <section className="rounded-[1.6rem] border border-dashed border-border/80 bg-card/70 p-5">
        <p className="text-sm font-semibold text-foreground">{emptyTitle}</p>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          {emptyDetail}
        </p>
      </section>
    );
  }

  return (
    <ul className="grid gap-3 md:grid-cols-2">
      {badges.map((badge) => {
        const tone = getBadgeTierTone(badge.tier);

        return (
          <li
            key={badge.code}
            className="rounded-[1.4rem] border border-border/80 bg-card/80 p-4"
          >
            <div className="flex flex-wrap items-center gap-2">
              <span
                className={`rounded-full px-3 py-1 text-xs font-medium uppercase tracking-[0.18em] ${tone.badge}`}
              >
                {tone.label}
              </span>
              <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                {badge.code}
              </span>
            </div>
            <h3 className="mt-3 text-lg font-semibold text-foreground">
              {badge.name}
            </h3>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              {badge.description}
            </p>
            <p className="mt-3 text-xs uppercase tracking-[0.18em] text-muted-foreground">
              Unlocked {formatUnixDate(badge.unlocked_at)}
            </p>
          </li>
        );
      })}
    </ul>
  );
}
