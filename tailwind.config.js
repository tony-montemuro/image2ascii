/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./static/**/*.js", "./templates/**/*.html"],
  theme: {
    extend: {
      animation: {
        popin: "popin 0.1s"
      },
      colors: {
        twitch: {
          DEFAULT: '#9146FF'
        },
        discord: {
          DEFAULT: '#5865F2'
        }
      },
      grayscale: {
        75: '75%'
      },
      keyframes: {
        popin: {
          "0%": {opacity: 0, transform: "scale(0.1)"},
          "100%": {opacity: 1, transform: "scale(1)"}
        }
      }
    },
  },
  plugins: [],
}

