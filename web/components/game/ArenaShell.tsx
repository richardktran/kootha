"use client";

import { motion } from "framer-motion";
import { ReactNode } from "react";

type ArenaShellProps = {
  children: ReactNode;
  className?: string;
};

export function ArenaShell({ children, className = "" }: ArenaShellProps) {
  return (
    <div className={`relative min-h-dvh overflow-hidden arena-bg ${className}`}>
      <div className="pointer-events-none absolute inset-0 arena-grid" aria-hidden />
      <div
        className="pointer-events-none absolute -left-24 top-24 h-64 w-64 rounded-full bg-pulse-lime/10 blur-3xl animate-float"
        aria-hidden
      />
      <div
        className="pointer-events-none absolute -right-16 bottom-32 h-72 w-72 rounded-full bg-pulse-coral/15 blur-3xl animate-float-delayed"
        aria-hidden
      />
      <div className="relative z-10">{children}</div>
    </div>
  );
}

export function BrandMark({ size = "md" }: { size?: "sm" | "md" | "lg" | "hero" }) {
  const sizes = {
    sm: "text-xl",
    md: "text-3xl",
    lg: "text-5xl md:text-6xl",
    hero: "text-6xl sm:text-7xl md:text-8xl",
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 12, scale: 0.96 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={{ duration: 0.45, ease: [0.22, 1, 0.36, 1] }}
      className="inline-flex flex-col items-start"
    >
      <span
        className={`text-brand leading-none text-[var(--pulse-lime)] drop-shadow-[0_4px_0_rgba(0,0,0,0.35)] ${sizes[size]}`}
      >
        Kootha
      </span>
      {size === "hero" || size === "lg" ? (
        <span className="mt-2 font-sans text-sm font-semibold uppercase tracking-[0.28em] text-[var(--muted)]">
          Live quiz arena
        </span>
      ) : null}
    </motion.div>
  );
}

export function StatusPill({
  children,
  tone = "default",
}: {
  children: ReactNode;
  tone?: "default" | "lime" | "coral" | "amber" | "sky";
}) {
  const tones = {
    default: "bg-white/10 text-white/80",
    lime: "bg-[var(--pulse-lime)] text-[var(--arena-ink)]",
    coral: "bg-[var(--pulse-coral)] text-white",
    amber: "bg-[var(--pulse-amber)] text-[var(--arena-ink)]",
    sky: "bg-[var(--pulse-sky)] text-[var(--arena-ink)]",
  };

  return <span className={`chip ${tones[tone]}`}>{children}</span>;
}
