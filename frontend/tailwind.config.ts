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
          50: "#FFF4ED",
          100: "#FFE4D4",
          200: "#FEC6A8",
          300: "#FD9E70",
          400: "#F9733A",
          500: "#E85D04",
          600: "#C2410C",
          700: "#9A3412",
          800: "#7C2D12",
          900: "#431407",
        },
        ink: {
          700: "#2C333A",
          800: "#1F2933",
          900: "#141B22",
        },
      },
      boxShadow: {
        card: "0 18px 40px -24px rgba(67, 20, 7, 0.35)",
      },
    },
  },
  plugins: [],
} satisfies Config;
