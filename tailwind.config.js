/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/template/**/*.html",
    "./web/src/**/*.js"
  ],
  theme: {
    extend: {
      colors: {
        ivory: "#FDFBF7",
        card: "#FFFFFF",
        ink: "#0B1B3D",
        mustard: "#F9AB06",
        mint: "#20C997",
        coral: "#FF4D4D",
        sand: "#F4EFE6"
      },
      boxShadow: {
        brutal: "4px 4px 0 #0B1B3D",
        press: "1px 1px 0 #0B1B3D"
      },
      fontFamily: {
        sans: ["Arial", "Helvetica", "sans-serif"],
        mono: ["ui-monospace", "SFMono-Regular", "Menlo", "monospace"]
      }
    }
  },
  plugins: []
};
