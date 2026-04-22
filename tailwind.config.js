/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    "./templates/**/*.templ",
    "./static/**/*.js"
  ],
  theme: {
    extend: {
      colors: {
        indigo: {
          50: '#EEF2FF',
          600: '#4F46E5',
          700: '#4338CA',
          800: '#3730A3',
        },
        teal: {
          400: '#2DD4BF',
          500: '#14B8A6',
        },
        slate: {
          50: '#F8FAFC',
          400: '#94A3B8',
          500: '#64748B',
          800: '#1E293B',
          900: '#0F172A',
        },
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
    },
  },
  plugins: [],
}
