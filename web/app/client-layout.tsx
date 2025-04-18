"use client";

import { Providers } from "./providers";
import { WebSocketProvider } from "./contexts/WebSocketContext";

export default function ClientLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <WebSocketProvider>
      <Providers themeProps={{ attribute: "class", defaultTheme: "dark" }}>
        <div className="relative flex flex-col min-h-screen w-screen">
          <main className="flex-grow w-full">{children}</main>
        </div>
      </Providers>
    </WebSocketProvider>
  );
}
