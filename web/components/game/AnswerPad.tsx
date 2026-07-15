"use client";

import { motion } from "framer-motion";

const OPTION_COLORS = [
  "var(--opt-a)",
  "var(--opt-b)",
  "var(--opt-c)",
  "var(--opt-d)",
] as const;

const OPTION_SHAPES = ["▲", "◆", "●", "■"] as const;

type AnswerPadProps = {
  options: string[];
  selectedAnswer: number | null;
  correctAnswer?: number;
  isRevealed: boolean;
  disabled: boolean;
  onSelect: (index: number) => void;
};

export function AnswerPad({
  options,
  selectedAnswer,
  correctAnswer,
  isRevealed,
  disabled,
  onSelect,
}: AnswerPadProps) {
  const stateFor = (index: number) => {
    if (!isRevealed || correctAnswer === undefined) return undefined;
    if (index === correctAnswer) return "correct";
    if (selectedAnswer === index) return "wrong";
    return "dim";
  };

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
      {options.map((option, index) => {
        const state = stateFor(index);
        const selected = selectedAnswer === index;

        return (
          <motion.button
            key={`${option}-${index}`}
            type="button"
            initial={{ opacity: 0, y: 14 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.06, duration: 0.35 }}
            disabled={disabled}
            data-selected={selected && !isRevealed}
            data-state={state}
            data-locked={disabled || isRevealed}
            className="answer-tile"
            style={{ background: OPTION_COLORS[index % OPTION_COLORS.length] }}
            onClick={() => onSelect(index)}
          >
            <span
              className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-black/20 text-lg"
              aria-hidden
            >
              {OPTION_SHAPES[index % OPTION_SHAPES.length]}
            </span>
            <span className="flex-1 leading-snug">{option}</span>
          </motion.button>
        );
      })}
    </div>
  );
}
