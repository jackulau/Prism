/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'primary': 'var(--color-editor-accent)',
        'editor-bg': 'var(--color-editor-bg)',
        'editor-surface': 'var(--color-editor-surface)',
        'editor-border': 'var(--color-editor-border)',
        'editor-text': 'var(--color-editor-text)',
        'editor-muted': 'var(--color-editor-muted)',
        'editor-accent': 'var(--color-editor-accent)',
        'editor-success': 'var(--color-editor-success)',
        'editor-warning': 'var(--color-editor-warning)',
        'editor-error': 'var(--color-editor-error)',
        'sidebar-bg': 'var(--color-sidebar-bg)',
        'sidebar-hover': 'var(--color-sidebar-hover)',
      },
      fontFamily: {
        mono: ['JetBrains Mono', 'Fira Code', 'Monaco', 'Consolas', 'monospace'],
      },
      animation: {
        'pulse-slow': 'pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'typing': 'typing 1s steps(3) infinite',
        'slide-in-right': 'slide-in-right 0.3s ease-out',
        'slide-in-down': 'slide-in-down 0.2s ease-out',
        'wiggle': 'wiggle 0.5s ease-in-out',
      },
      keyframes: {
        typing: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0' },
        },
        'slide-in-right': {
          '0%': { transform: 'translateX(100%)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        'slide-in-down': {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        'wiggle': {
          '0%, 100%': { transform: 'rotate(0deg)' },
          '25%': { transform: 'rotate(-10deg)' },
          '50%': { transform: 'rotate(10deg)' },
          '75%': { transform: 'rotate(-5deg)' },
        },
      },
    },
  },
  plugins: [],
}
