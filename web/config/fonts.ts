import {
  Bricolage_Grotesque as FontDisplay,
  Nunito as FontSans,
  JetBrains_Mono as FontMono,
} from "next/font/google";

export const fontSans = FontSans({
  subsets: ["latin"],
  variable: "--font-sans",
  weight: ["400", "600", "700", "800"],
});

export const fontDisplay = FontDisplay({
  subsets: ["latin"],
  variable: "--font-display",
  weight: ["600", "700", "800"],
});

export const fontMono = FontMono({
  subsets: ["latin"],
  variable: "--font-mono",
});
