/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./static/**/*.{html,js}"],
  theme: {
    extend: {
      colors: {
        twitch: {
          DEFAULT: '#9146FF'
        },
        discord: {
          DEFAULT: '#5865F2'
        }
      }
    },
  },
  plugins: [],
}

