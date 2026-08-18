import type { Config } from "tailwindcss";

export default {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        mahak: {
          50: "#FDE7F0",
          100: "#F8C5D8",
          200: "#F48FB1",
          300: "#EC5C8E",
          400: "#E91E63",
          500: "#D81B60",
          600: "#C2185B",
          700: "#AD1457",
          800: "#880E4F",
          900: "#4A062B",
        },
        ink: {
          700: "#2C333A",
          800: "#1F2933",
          900: "#141B22",
        },
      },
      boxShadow: {
        card: "0 18px 40px -24px rgba(216, 27, 96, 0.35)",
      },
    },
  },
  plugins: [],
} satisfies Config;
