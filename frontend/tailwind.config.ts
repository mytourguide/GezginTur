import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: { 50: "#effaf6", 500: "#0d9488", 600: "#0f766e", 700: "#115e59", 900: "#134e4a" },
      },
    },
  },
  plugins: [],
};
export default config;
