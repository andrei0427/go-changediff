module.exports = {
  content: ["./web/views/**/*.html", "!./web/views/widget/**/*.html"],
  theme: {
    extend: {
      colors: {
        green: "#42C671",
      },
    },
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
