module.exports = {
  content: ["./web/views/**/*.html"],
  theme: {
    extend: {
      colors: {
        blue: "#0565AB",
      },
    },
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
