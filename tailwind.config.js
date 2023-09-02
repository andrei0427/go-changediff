module.exports = {
  content: ["./web/views/**/*.html", "!./web/views/widget/**/*.html"],
  theme: {
    extend: {
      colors: {
        orange: "#f67c4c",
      },
    },
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
