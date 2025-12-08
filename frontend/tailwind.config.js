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
      },
      keyframes: {
        typing: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0' },
        },
      },
    },
  },
  plugins: [],
}
