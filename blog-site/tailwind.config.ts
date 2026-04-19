import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Space Grotesk", "system-ui", "-apple-system", "sans-serif"],
        mono: ["JetBrains Mono", "Fira Code", "monospace"],
      },
      colors: {
        cream: "#FFFDF5",
        ink: "#0A0A0A",
        yellow: "#FFD93D",
        red: "#FF4A4A",
        green: "#05D98F",
        background: "var(--background)",
        foreground: "var(--foreground)",
      },
      boxShadow: {
        brutal: "4px 4px 0px #0A0A0A",
        "brutal-sm": "2px 2px 0px #0A0A0A",
        "brutal-lg": "6px 6px 0px #0A0A0A",
        "brutal-xl": "8px 8px 0px #0A0A0A",
        "brutal-yellow": "4px 4px 0px #FFD93D",
        "brutal-red": "4px 4px 0px #FF4A4A",
        "brutal-green": "4px 4px 0px #05D98F",
      },
      borderWidth: {
        "3": "3px",
      },
    },
  },
  plugins: [require("@tailwindcss/typography")],
};
export default config;
