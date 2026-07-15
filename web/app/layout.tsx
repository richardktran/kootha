import "@/styles/globals.css";
import { Metadata, Viewport } from "next";
import clsx from "clsx";

import ClientLayout from "./client-layout";

import { fontDisplay, fontSans } from "@/config/fonts";
import { siteConfig } from "@/config/site";

export const metadata: Metadata = {
  title: {
    default: siteConfig.name,
    template: `%s · ${siteConfig.name}`,
  },
  description: siteConfig.description,
  icons: {
    icon: "/favicon.ico",
  },
};

export const viewport: Viewport = {
  themeColor: "#071a17",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html suppressHydrationWarning lang="en">
      <body
        className={clsx(
          "min-h-dvh font-sans antialiased",
          fontSans.variable,
          fontDisplay.variable,
        )}
      >
        <ClientLayout>{children}</ClientLayout>
      </body>
    </html>
  );
}
