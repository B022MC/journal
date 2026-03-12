import type { Metadata } from "next";
import { SiteFooter } from "@/components/layout/site-footer";
import { SiteHeader } from "@/components/layout/site-header";
import { getApiRuntimeSnapshot } from "@/lib/api/runtime";
import "./globals.css";

export const metadata: Metadata = {
  title: {
    default: "S.H.I.T Journal",
    template: "%s | S.H.I.T Journal",
  },
  description:
    "Archive Lab shell for paper reading, search, and community governance.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const runtime = getApiRuntimeSnapshot();

  return (
    <html lang="en">
      <body className="min-h-screen bg-background font-sans text-foreground antialiased">
        <div className="relative flex min-h-screen flex-col">
          <SiteHeader runtime={runtime} />
          <main className="flex-1">{children}</main>
          <SiteFooter runtime={runtime} />
        </div>
      </body>
    </html>
  );
}
