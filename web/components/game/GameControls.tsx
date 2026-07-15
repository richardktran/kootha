"use client";

import { ButtonHTMLAttributes, InputHTMLAttributes, forwardRef } from "react";

type ButtonVariant = "primary" | "secondary" | "ghost" | "host";

const variantClass: Record<ButtonVariant, string> = {
  primary: "btn-primary",
  secondary: "btn-secondary",
  ghost: "btn-ghost",
  host: "btn-host",
};

type GameButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  fullWidth?: boolean;
};

export function GameButton({
  variant = "primary",
  fullWidth,
  className = "",
  children,
  ...props
}: GameButtonProps) {
  return (
    <button
      type="button"
      className={`${variantClass[variant]} ${fullWidth ? "w-full" : ""} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}

type GameInputProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  light?: boolean;
  error?: string | null;
};

export const GameInput = forwardRef<HTMLInputElement, GameInputProps>(
  function GameInput(
    { label, light, error, id, className = "", ...props },
    ref,
  ) {
    return (
      <div className="space-y-2">
        {label ? (
          <label
            htmlFor={id}
            className={`block text-sm font-bold uppercase tracking-wider ${
              light ? "text-black/60" : "text-[var(--muted)]"
            }`}
          >
            {label}
          </label>
        ) : null}
        <input
          ref={ref}
          id={id}
          className={`${light ? "input-game-light" : "input-game"} ${className}`}
          {...props}
        />
        {error ? (
          <p className="text-sm font-semibold text-[var(--pulse-coral)]">{error}</p>
        ) : null}
      </div>
    );
  },
);
