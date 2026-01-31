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
        'progress-shimmer': 'progress-shimmer 1.5s ease-in-out infinite',
        'progress-stripe': 'progress-stripe 1s linear infinite',
        'bounce-dot': 'bounce-dot 1.4s ease-in-out infinite',
        'wave': 'wave 1.2s ease-in-out infinite',
        'typewriter-blink': 'typewriter-blink 1s step-end infinite',
      },
      keyframes: {
        typing: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0' },
        },
        'progress-shimmer': {
          '0%': { opacity: '0.6', transform: 'translateX(-100%)' },
          '50%': { opacity: '1' },
          '100%': { opacity: '0.6', transform: 'translateX(100%)' },
        },
        'progress-stripe': {
          'from': { backgroundPosition: '1rem 0' },
          'to': { backgroundPosition: '0 0' },
        },
        'bounce-dot': {
          '0%, 80%, 100%': { transform: 'translateY(0)' },
          '40%': { transform: 'translateY(-50%)' },
        },
        'wave': {
          '0%, 100%': { height: '25%' },
          '50%': { height: '100%' },
        },
        'typewriter-blink': {
          '0%, 50%': { opacity: '1' },
          '51%, 100%': { opacity: '0' },
        },
      },
      backgroundImage: {
        'stripes': 'linear-gradient(45deg, rgba(255,255,255,0.15) 25%, transparent 25%, transparent 50%, rgba(255,255,255,0.15) 50%, rgba(255,255,255,0.15) 75%, transparent 75%, transparent)',
      },
      backgroundSize: {
        'stripes': '1rem 1rem',
      },
    },
  },
  plugins: [],
}
