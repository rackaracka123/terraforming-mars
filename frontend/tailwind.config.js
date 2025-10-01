/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Base Dark Theme Colors
        'space-black': {
          DEFAULT: '#0a0a0f',
          darker: '#050509',
          light: '#141420',
        },
        'space-blue': {
          50: 'rgba(30, 60, 150, 0.05)',
          100: 'rgba(30, 60, 150, 0.1)',
          200: 'rgba(30, 60, 150, 0.2)',
          300: 'rgba(30, 60, 150, 0.3)',
          400: 'rgba(30, 60, 150, 0.4)',
          500: 'rgba(30, 60, 150, 0.5)',
          600: 'rgba(30, 60, 150, 0.6)',
          700: 'rgba(30, 60, 150, 0.7)',
          800: 'rgba(30, 60, 150, 0.8)',
          900: 'rgba(30, 60, 150, 0.9)',
          solid: 'rgb(30, 60, 150)',
        },
        'glow-blue': {
          DEFAULT: 'rgba(30, 60, 150, 0.6)',
          light: 'rgba(30, 60, 150, 0.4)',
          strong: 'rgba(30, 60, 150, 0.9)',
        },
        'cyber-cyan': {
          DEFAULT: 'rgba(0, 212, 255, 0.3)',
          light: 'rgba(0, 212, 255, 0.2)',
          strong: 'rgba(0, 212, 255, 0.5)',
        },

        // Game Resource Colors (Must be exact)
        'mars-red': '#e74c3c',
        'plant-green': '#27ae60',
        'credit-yellow': '#f1c40f',
        'steel-gray': '#95a5a6',
        'heat-orange': '#e67e22',
        'energy-blue': '#3498db',

        // UI State Colors
        'error-red': {
          DEFAULT: '#ff6b6b',
          light: '#ffcdd2',
          glow: 'rgba(244, 67, 54, 0.7)',
        },
        'success-green': {
          DEFAULT: '#4caf50',
          light: '#c8e6c9',
        },

        // Admin & Special
        'admin-purple': {
          300: 'rgba(155, 89, 182, 0.3)',
          500: 'rgba(155, 89, 182, 0.5)',
          600: 'rgba(155, 89, 182, 0.6)',
          800: 'rgba(155, 89, 182, 0.8)',
        },
        'effect-brown': {
          light: 'rgba(160, 110, 60, 0.4)',
          DEFAULT: 'rgba(139, 89, 42, 0.35)',
        },
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', '"Segoe UI"', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', '"Fira Sans"', '"Droid Sans"', '"Helvetica Neue"', 'sans-serif'],
        orbitron: ['Orbitron', 'sans-serif'],
        mono: ['source-code-pro', 'Menlo', 'Monaco', 'Consolas', '"Courier New"', 'monospace'],
      },
      backdropBlur: {
        'space': '10px',
        'space-light': '5px',
      },
      boxShadow: {
        'glow': '0 0 30px rgba(30, 60, 150, 0.6)',
        'glow-strong': '0 0 30px rgba(30, 60, 150, 0.8)',
        'glow-sm': '0 0 20px rgba(30, 60, 150, 0.4)',
        'glow-lg': '0 0 60px rgba(30, 60, 150, 0.3)',
        'card': '0 6px 24px rgba(0, 0, 0, 0.5), 0 2px 20px rgba(0, 212, 255, 0.2)',
        'card-hover': '0 6px 24px rgba(0, 0, 0, 0.5), 0 2px 20px rgba(0, 212, 255, 0.3)',
      },
      dropShadow: {
        'icon': '0 1px 2px rgba(0, 0, 0, 0.5)',
        'icon-strong': '0 1px 3px rgba(0, 0, 0, 0.7)',
        'red-glow': ['0 1px 2px rgba(0, 0, 0, 0.5)', '0 0 1px rgba(244, 67, 54, 0.9)', '0 0 2px rgba(244, 67, 54, 0.7)'],
      },
      textShadow: {
        'glow': '0 0 30px rgba(30, 60, 150, 0.6)',
        'glow-strong': '0 0 30px rgba(30, 60, 150, 0.8)',
        'glow-sm': '0 0 20px rgba(30, 60, 150, 0.4)',
        'dark': '1px 1px 2px rgba(0, 0, 0, 0.7)',
      },
      animation: {
        'pulse-slow': 'pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
      letterSpacing: {
        'wider-xl': '0.15em',
        'wider-2xl': '0.2em',
      },
    },
  },
  plugins: [
    // Plugin for text-shadow support
    function({ matchUtilities, theme }) {
      matchUtilities(
        {
          'text-shadow': (value) => ({
            textShadow: value,
          }),
        },
        { values: theme('textShadow') }
      )
    },
  ],
}
