module.exports = {
  content: [
    "./web/views/**/*.html",
    "!./web/views/index.html",
    "!./web/views/widget/**/*.html",
  ],
  theme: {
    extend: {
      colors: {
        emerald: "#42C671",
      },
    },
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
