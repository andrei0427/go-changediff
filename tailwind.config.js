module.exports = {
  content: ["./views/**/*.html"],
  theme: {
    extend: {
      colors: {
        blue: "#0565AB",
      },
    },
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
