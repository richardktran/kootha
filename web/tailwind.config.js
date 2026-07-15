import { nextui } from "@nextui-org/theme";

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./node_modules/@nextui-org/theme/dist/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ["var(--font-sans)", "system-ui", "sans-serif"],
        display: ["var(--font-display)", "var(--font-sans)", "system-ui"],
        mono: ["var(--font-mono)", "monospace"],
      },
      colors: {
        arena: {
          ink: "var(--arena-ink)",
          deep: "var(--arena-deep)",
          mid: "var(--arena-mid)",
          fog: "var(--arena-fog)",
        },
        pulse: {
          lime: "var(--pulse-lime)",
          coral: "var(--pulse-coral)",
          amber: "var(--pulse-amber)",
          sky: "var(--pulse-sky)",
        },
      },
      boxShadow: {
        pop: "var(--shadow-pop)",
        press: "var(--shadow-press)",
      },
      keyframes: {
        "fade-up": {
          "0%": { opacity: "0", transform: "translateY(16px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        "pop-in": {
          "0%": { opacity: "0", transform: "scale(0.88)" },
          "70%": { transform: "scale(1.04)" },
          "100%": { opacity: "1", transform: "scale(1)" },
        },
        "slide-in": {
          "0%": { opacity: "0", transform: "translateX(-12px)" },
          "100%": { opacity: "1", transform: "translateX(0)" },
        },
      },
      animation: {
        "fade-up": "fade-up 0.45s ease-out both",
        "pop-in": "pop-in 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both",
        "slide-in": "slide-in 0.35s ease-out both",
      },
    },
  },
  darkMode: "class",
  plugins: [nextui()],
};
