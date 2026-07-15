"use client";

import { motion } from "framer-motion";

type TimerRingProps = {
  timeLeft: number;
  timeLimit: number;
};

export function TimerRing({ timeLeft, timeLimit }: TimerRingProps) {
  const ratio = Math.max(0, Math.min(1, timeLeft / Math.max(timeLimit, 1)));
  const urgency = timeLeft <= 5;
  const r = 54;
  const c = 2 * Math.PI * r;
  const offset = c * (1 - ratio);

  return (
    <div className="relative mx-auto flex h-28 w-28 items-center justify-center">
      <svg className="absolute inset-0 -rotate-90" viewBox="0 0 120 120" aria-hidden>
        <circle
          cx="60"
          cy="60"
          r={r}
          fill="none"
          stroke="rgba(255,255,255,0.12)"
          strokeWidth="10"
        />
        <motion.circle
          cx="60"
          cy="60"
          r={r}
          fill="none"
          stroke={urgency ? "var(--pulse-coral)" : "var(--pulse-lime)"}
          strokeWidth="10"
          strokeLinecap="round"
          strokeDasharray={c}
          animate={{ strokeDashoffset: offset }}
          transition={{ duration: 0.35, ease: "easeOut" }}
        />
      </svg>
      <div
        className={`text-display text-4xl tabular-nums ${
          urgency
            ? "animate-urgency text-[var(--pulse-coral)]"
            : "text-[var(--pulse-lime)]"
        }`}
      >
        {timeLeft}
      </div>
    </div>
  );
}
