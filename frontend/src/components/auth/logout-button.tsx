"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { clearBrowserSession } from "@/lib/auth/browser-session";
import { cn } from "@/lib/utils";

export function LogoutButton({
  returnTo = "/",
  className,
}: {
  returnTo?: string;
  className?: string;
}) {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  return (
    <div className="space-y-2">
      <button
        type="button"
        disabled={isPending}
        className={cn(
          "inline-flex rounded-full border border-border/80 bg-background/80 px-4 py-2 text-sm font-medium text-foreground disabled:opacity-60",
          className,
        )}
        onClick={() => {
          setError(null);

          startTransition(async () => {
            const result = await clearBrowserSession();
            if (!result.ok) {
              setError(result.message);
              return;
            }

            router.push(
              `/login?reason=signed_out&returnTo=${encodeURIComponent(returnTo)}`,
            );
            router.refresh();
          });
        }}
      >
        {isPending ? "Signing out…" : "Sign out"}
      </button>
      {error ? <p className="text-xs text-[#8b312e]">{error}</p> : null}
    </div>
  );
}
