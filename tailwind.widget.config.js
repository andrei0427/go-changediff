module.exports = {
  content: ["./web/views/widget/**/*.html"],
  theme: {
    extend: {
      colors: {},
    },
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
