/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./static/**/*.js', './templates/**/*.html'],
  darkMode: 'class',
  theme: {
    extend: {
      animation: {
        popin: 'popin 0.1s',
        popout: 'popout 0.1s'
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
          '0%': {opacity: 0, transform: 'scale(0.1)'},
          '100%': {opacity: 1, transform: 'scale(1)'}
        },
        popout: {
          '0%': {opacity: 1, transform: 'scale(1)'},
          '100%': {opacity: 0, transform: 'scale(0.1)'}
        }
      }
    },
  },
  plugins: [],
}

