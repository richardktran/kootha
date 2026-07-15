"use client";

import { motion } from "framer-motion";

import { Participant } from "@/types/quiz";
import { StatusPill } from "@/components/game/ArenaShell";

type PlayerRosterProps = {
  participants: Participant[];
  hostId: string;
  currentUserId: string;
  compact?: boolean;
};

const AVATAR_TONES = [
  "bg-[var(--opt-a)]",
  "bg-[var(--opt-b)]",
  "bg-[var(--opt-c)]",
  "bg-[var(--opt-d)]",
  "bg-[var(--pulse-coral)]",
  "bg-[var(--pulse-sky)]",
];

function avatarTone(name: string) {
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash = (hash + name.charCodeAt(i) * 17) % 6;
  return AVATAR_TONES[hash];
}

export function PlayerRoster({
  participants,
  hostId,
  currentUserId,
  compact,
}: PlayerRosterProps) {
  return (
    <div className={compact ? "flex flex-wrap gap-2" : "space-y-2"}>
      {participants.map((p, i) => {
        if (compact) {
          return (
            <motion.div
              key={p.id}
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: i * 0.04 }}
              className="inline-flex items-center gap-2 rounded-full border border-[var(--line)] bg-[var(--surface)] py-1.5 pl-1.5 pr-3"
            >
              <span
                className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-black text-white ${avatarTone(p.name)}`}
              >
                {p.name.slice(0, 1).toUpperCase()}
              </span>
              <span className="text-sm font-bold">{p.name}</span>
            </motion.div>
          );
        }

        return (
          <motion.div
            key={p.id}
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: i * 0.05 }}
            className="flex items-center justify-between rounded-[var(--radius-tile)] border border-[var(--line)] bg-[var(--surface)] px-3 py-3"
          >
            <div className="flex min-w-0 items-center gap-3">
              <span
                className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-sm font-black text-white ${avatarTone(p.name)}`}
              >
                {p.name.slice(0, 1).toUpperCase()}
              </span>
              <div className="min-w-0">
                <div className="truncate font-bold">{p.name}</div>
                <div className="mt-1 flex flex-wrap gap-1.5">
                  {p.id === hostId ? <StatusPill tone="amber">Host</StatusPill> : null}
                  {p.id === currentUserId ? <StatusPill tone="sky">You</StatusPill> : null}
                </div>
              </div>
            </div>
            <div className="text-right">
              <div className="text-display text-xl text-[var(--pulse-lime)] tabular-nums">
                {p.score}
              </div>
              <div className="text-[10px] font-bold uppercase tracking-wider text-[var(--muted)]">
                pts
              </div>
            </div>
          </motion.div>
        );
      })}
    </div>
  );
}

type LeaderboardProps = {
  participants: Participant[];
  currentUserId: string;
  title?: string;
  finished?: boolean;
};

export function Leaderboard({
  participants,
  currentUserId,
  title = "Leaderboard",
  finished,
}: LeaderboardProps) {
  const ranked = [...participants].sort((a, b) => b.score - a.score);
  const podiumColors = [
    "from-[#ffd700]/30 to-transparent border-[#ffd700]/50",
    "from-[#c0c0c0]/25 to-transparent border-white/30",
    "from-[#cd7f32]/25 to-transparent border-[#cd7f32]/40",
  ];

  return (
    <div className="space-y-5">
      <div className="text-center">
        <h2 className="text-display text-3xl text-[var(--pulse-lime)] sm:text-4xl">
          {finished ? "Final Results" : title}
        </h2>
        {finished ? (
          <p className="mt-2 text-[var(--muted)]">
            {ranked[0]
              ? `${ranked[0].name} takes the crown`
              : "Thanks for playing"}
          </p>
        ) : null}
      </div>

      {finished && ranked.length > 0 ? (
        <div className="mx-auto flex max-w-md items-end justify-center gap-2 pt-2">
          {[ranked[1], ranked[0], ranked[2]].map((p, visualIndex) => {
            if (!p) return <div key={visualIndex} className="w-24" />;
            const place = visualIndex === 1 ? 1 : visualIndex === 0 ? 2 : 3;
            const heights = ["h-24", "h-32", "h-20"];
            return (
              <motion.div
                key={p.id}
                initial={{ opacity: 0, y: 24 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 * place, type: "spring", stiffness: 200 }}
                className="flex w-28 flex-col items-center"
              >
                <div
                  className={`mb-2 flex h-10 w-10 items-center justify-center rounded-full text-sm font-black text-white ${avatarTone(p.name)}`}
                >
                  {p.name.slice(0, 1).toUpperCase()}
                </div>
                <div className="mb-1 truncate text-center text-sm font-bold">{p.name}</div>
                <div className="mb-2 text-display text-[var(--pulse-lime)]">{p.score}</div>
                <div
                  className={`flex w-full items-start justify-center rounded-t-xl border bg-gradient-to-b pt-2 text-display text-2xl ${heights[visualIndex]} ${podiumColors[place - 1]}`}
                >
                  {place}
                </div>
              </motion.div>
            );
          })}
        </div>
      ) : null}

      <div className="space-y-2">
        {ranked.map((p, index) => {
          const isYou = p.id === currentUserId;
          return (
            <motion.div
              key={p.id}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
              className={`flex items-center justify-between rounded-[var(--radius-tile)] border px-4 py-3 ${
                isYou
                  ? "border-[var(--pulse-lime)]/50 bg-[var(--pulse-lime)]/10"
                  : "border-[var(--line)] bg-[var(--surface)]"
              }`}
            >
              <div className="flex items-center gap-3">
                <span
                  className={`flex h-9 w-9 items-center justify-center rounded-lg text-sm font-black ${
                    index === 0
                      ? "bg-[var(--pulse-amber)] text-[var(--arena-ink)]"
                      : "bg-white/10 text-white/70"
                  }`}
                >
                  {index + 1}
                </span>
                <div>
                  <div className="font-bold">
                    {p.name}
                    {isYou ? (
                      <span className="ml-2 text-xs text-[var(--pulse-lime)]">you</span>
                    ) : null}
                  </div>
                </div>
              </div>
              <div className="text-display text-xl tabular-nums text-[var(--pulse-lime)]">
                {p.score}
              </div>
            </motion.div>
          );
        })}
      </div>
    </div>
  );
}

type FeedbackBannerProps = {
  kind: "correct" | "incorrect" | "timeout";
};

export function FeedbackBanner({ kind }: FeedbackBannerProps) {
  const copy = {
    correct: { title: "Nice hit!", sub: "Points locked in", tone: "bg-[var(--opt-d)]" },
    incorrect: { title: "Missed it", sub: "Next round is yours", tone: "bg-[var(--opt-a)]" },
    timeout: { title: "Time's up", sub: "No answer this round", tone: "bg-[var(--pulse-amber)] text-[var(--arena-ink)]" },
  }[kind];

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.92 }}
      animate={{ opacity: 1, scale: 1 }}
      className={`rounded-[var(--radius-tile)] px-5 py-4 text-center shadow-pop ${copy.tone}`}
    >
      <div className="text-display text-2xl sm:text-3xl">{copy.title}</div>
      <div className="mt-1 text-sm font-semibold opacity-90">{copy.sub}</div>
    </motion.div>
  );
}
