"use client";

import { Providers } from "./providers";
import { UserProvider } from "./contexts/UserContext";
import { WebSocketProvider } from "./contexts/WebSocketContext";

export default function ClientLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <UserProvider>
      <WebSocketProvider>
        <Providers themeProps={{ attribute: "class", defaultTheme: "dark" }}>
          <div className="relative flex min-h-dvh w-full flex-col">
            {children}
          </div>
        </Providers>
      </WebSocketProvider>
    </UserProvider>
  );
}
