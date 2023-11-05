module.exports = {
  content: ["./web/views/widget/**/*.html"],
  theme: {
    extend: {
      colors: {
        transparent: "transparent",
        "white-100": "rgba(255,255,255,1)",
      },
    },
    backgroundImage: (theme) => ({
      "gradient-to-b": "linear-gradient(to bottom, var(--tw-gradient-stops))",
    }),
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
