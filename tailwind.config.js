/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./static/**/*.js", "./templates/**/*.html"],
  theme: {
    extend: {
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
      }
    },
  },
  plugins: [],
}

