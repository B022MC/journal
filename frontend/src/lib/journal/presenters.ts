import type { PaperItem } from "@/lib/journal/contracts";

const zoneToneMap: Record<
  string,
  { badge: string; border: string; label: string }
> = {
  latrine: {
    badge: "bg-[#426b54]/12 text-[#426b54]",
    border: "border-[#426b54]/20",
    label: "Latrine",
  },
  septic_tank: {
    badge: "bg-[#8a4b2a]/12 text-[#8a4b2a]",
    border: "border-[#8a4b2a]/20",
    label: "Septic Tank",
  },
  stone: {
    badge: "bg-[#53606d]/12 text-[#53606d]",
    border: "border-[#53606d]/20",
    label: "Stone",
  },
  sediment: {
    badge: "bg-[#6b6a3c]/12 text-[#5a5a2d]",
    border: "border-[#6b6a3c]/20",
    label: "Sediment",
  },
};

const badgeTierToneMap: Record<string, { badge: string; label: string }> = {
  bronze: {
    badge: "bg-[#8a4b2a]/12 text-[#8a4b2a]",
    label: "Bronze",
  },
  silver: {
    badge: "bg-[#53606d]/12 text-[#53606d]",
    label: "Silver",
  },
  gold: {
    badge: "bg-[#a17a20]/12 text-[#8a6616]",
    label: "Gold",
  },
  platinum: {
    badge: "bg-[#426b54]/12 text-[#426b54]",
    label: "Platinum",
  },
};

const roleLabelMap: Record<number, string> = {
  0: "Member",
  1: "Scooper",
  2: "Editor",
  3: "Admin",
};

export function getZoneTone(zone: string) {
  return zoneToneMap[zone.toLowerCase()] ?? {
    badge: "bg-secondary/70 text-foreground",
    border: "border-border/70",
    label: zone || "Unclassified",
  };
}

export function splitKeywords(keywords: string) {
  return keywords
    .split(/[;,]/)
    .map((keyword) => keyword.trim())
    .filter(Boolean)
    .slice(0, 6);
}

export function excerpt(text: string, maxLength = 180) {
  const normalized = text.replace(/\s+/g, " ").trim();

  if (normalized.length <= maxLength) {
    return normalized;
  }

  return `${normalized.slice(0, maxLength - 1).trimEnd()}…`;
}

export function formatScore(value: number) {
  return Number.isFinite(value) ? value.toFixed(1) : "0.0";
}

export function formatUnixDate(timestamp: number) {
  if (!timestamp) {
    return "Unknown date";
  }

  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
  }).format(new Date(timestamp * 1000));
}

export function formatCompactNumber(value: number) {
  return new Intl.NumberFormat("en", {
    maximumFractionDigits: 1,
    notation: "compact",
  }).format(value);
}

export function formatRoleLabel(role: number) {
  return roleLabelMap[role] ?? `Role ${role}`;
}

export function getBadgeTierTone(tier: string) {
  const normalized = tier.trim().toLowerCase();
  return badgeTierToneMap[normalized] ?? {
    badge: "bg-secondary/70 text-foreground",
    label: tier || "Badge",
  };
}

export function buildZoneSummary(papers: PaperItem[]) {
  const summary = new Map<
    string,
    { zone: string; count: number; latestTitle: string | null }
  >();

  for (const paper of papers) {
    const existing = summary.get(paper.zone) ?? {
      zone: paper.zone,
      count: 0,
      latestTitle: null,
    };

    summary.set(paper.zone, {
      zone: paper.zone,
      count: existing.count + 1,
      latestTitle: existing.latestTitle ?? paper.title,
    });
  }

  return Array.from(summary.values());
}

export function parseSearchParam(
  value: string | string[] | undefined,
  fallback = "",
) {
  if (Array.isArray(value)) {
    return value[0] ?? fallback;
  }

  return value ?? fallback;
}
