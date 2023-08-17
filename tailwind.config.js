module.exports = {
  content: ["./web/views/**/*.html"],
  theme: {
    extend: {
      colors: {
        orange: "#f67c4c",
      },
    },
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
