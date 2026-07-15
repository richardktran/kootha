"use client";

import { useEffect } from "react";

import { ArenaShell, BrandMark, GameButton } from "@/components/game";

export default function Error({
  error,
  reset,
}: {
  error: Error;
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <ArenaShell>
      <div className="flex min-h-dvh flex-col items-center justify-center gap-6 px-4 text-center">
        <BrandMark size="md" />
        <h2 className="text-display text-3xl">Something went wrong</h2>
        <p className="max-w-sm text-[var(--muted)]">
          The arena hit a snag. Try again — your room may still be waiting.
        </p>
        <GameButton onClick={() => reset()}>Try again</GameButton>
      </div>
    </ArenaShell>
  );
}
