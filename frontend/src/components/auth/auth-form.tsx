"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { apiFetch } from "@/lib/api";
import { writeBrowserToken } from "@/lib/auth/browser-session";
import type { IdResponse, LoginResponse } from "@/lib/journal/contracts";

export function AuthForm({
  mode,
  registered = false,
  returnTo = "/",
}: {
  mode: "login" | "register";
  registered?: boolean;
  returnTo?: string;
}) {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(
    registered && mode === "login"
      ? "Account created. Sign in to continue."
      : null,
  );
  const [isPending, startTransition] = useTransition();

  return (
    <form
      className="space-y-4"
      onSubmit={(event) => {
        event.preventDefault();
        const form = event.currentTarget;
        const formData = new FormData(form);

        setError(null);
        setMessage(null);

        startTransition(async () => {
          if (mode === "login") {
            const result = await apiFetch<LoginResponse>("/user/login", {
              body: {
                username: String(formData.get("username") ?? ""),
                password: String(formData.get("password") ?? ""),
              },
              method: "POST",
            });

            if (!result.ok) {
              setError(result.error.detail);
              return;
            }

            writeBrowserToken(result.data.token);
            setMessage("Login succeeded. Redirecting to the requested page.");
            router.push(returnTo);
            router.refresh();
            return;
          }

          const registerPayload = {
            username: String(formData.get("username") ?? ""),
            email: String(formData.get("email") ?? ""),
            nickname: String(formData.get("nickname") ?? ""),
            password: String(formData.get("password") ?? ""),
          } satisfies Record<string, string>;

          const result = await apiFetch<IdResponse>("/user/register", {
            body: registerPayload,
            method: "POST",
          });

          if (!result.ok) {
            setError(result.error.detail);
            return;
          }

          setMessage("Registration succeeded. Redirecting to sign-in.");
          router.push(`/login?registered=1&returnTo=${encodeURIComponent(returnTo)}`);
        });
      }}
    >
      <label className="block text-sm text-muted-foreground">
        Username
        <input
          required
          name="username"
          className="mt-2 w-full rounded-[1.25rem] border border-border/80 bg-card px-3 py-3 text-foreground"
          placeholder="reader_name"
        />
      </label>

      {mode === "register" ? (
        <>
          <label className="block text-sm text-muted-foreground">
            Email
            <input
              required
              name="email"
              type="email"
              className="mt-2 w-full rounded-[1.25rem] border border-border/80 bg-card px-3 py-3 text-foreground"
              placeholder="you@example.com"
            />
          </label>
          <label className="block text-sm text-muted-foreground">
            Nickname
            <input
              name="nickname"
              className="mt-2 w-full rounded-[1.25rem] border border-border/80 bg-card px-3 py-3 text-foreground"
              placeholder="Optional display name"
            />
          </label>
        </>
      ) : null}

      <label className="block text-sm text-muted-foreground">
        Password
        <input
          required
          minLength={6}
          name="password"
          type="password"
          className="mt-2 w-full rounded-[1.25rem] border border-border/80 bg-card px-3 py-3 text-foreground"
          placeholder="Minimum 6 characters"
        />
      </label>

      {error ? <p className="text-sm text-[#8b312e]">{error}</p> : null}
      {message ? <p className="text-sm text-[#426b54]">{message}</p> : null}

      <button
        type="submit"
        disabled={isPending}
        className="inline-flex rounded-full bg-primary px-5 py-3 text-sm font-medium text-primary-foreground disabled:opacity-60"
      >
        {isPending
          ? mode === "login"
            ? "Signing in…"
            : "Creating account…"
          : mode === "login"
            ? "Sign in"
            : "Create account"}
      </button>
    </form>
  );
}
