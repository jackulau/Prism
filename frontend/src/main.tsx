import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import { ErrorBoundary } from './components/ErrorBoundary'
import { AuthGuard } from './components/auth/AuthGuard'
import { AuthPage } from './components/auth/AuthPage'
import './index.css'

// PWA Service Worker Registration
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js', { scope: '/' })
      .catch(() => {
        // Service worker registration failed - PWA features will be unavailable
      })
  })
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <AuthGuard fallback={<AuthPage />}>
        <App />
      </AuthGuard>
    </ErrorBoundary>
  </React.StrictMode>,
)
