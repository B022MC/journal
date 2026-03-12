import { redirect } from "next/navigation";
import { getServerAuthSession } from "@/lib/api/server";
import type { ApiDiagnostic } from "@/lib/api/errors";
import type { UserInfo } from "@/lib/journal/contracts";
import { getCurrentUser } from "@/lib/journal/server";

export type RequiredUserResult =
  | { ok: true; user: UserInfo }
  | { ok: false; error: ApiDiagnostic };

export async function requireCurrentUser(
  returnTo: string,
): Promise<RequiredUserResult> {
  const session = await getServerAuthSession();
  if (!session.token) {
    redirect(`/login?reason=protected&returnTo=${encodeURIComponent(returnTo)}`);
  }

  const result = await getCurrentUser();
  if (result.ok) {
    return { ok: true, user: result.data };
  }

  if (
    result.error.kind === "auth" ||
    result.error.code === "AUTH_TOKEN_MISSING"
  ) {
    redirect(
      `/login?reason=protected&returnTo=${encodeURIComponent(returnTo)}`,
    );
  }

  return { ok: false, error: result.error };
}
