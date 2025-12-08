import { useState } from 'react';
import { LoginForm } from './LoginForm';
import { RegisterForm } from './RegisterForm';

export function AuthPage() {
  const [mode, setMode] = useState<'login' | 'register'>('login');

  return (
    <div className="min-h-screen bg-editor-bg flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-editor-text mb-2">Prism</h1>
          <p className="text-editor-muted">AI-powered code assistant</p>
        </div>

        <div className="bg-editor-surface border border-editor-border rounded-xl p-6">
          {mode === 'login' ? (
            <LoginForm
              onSuccess={() => {}}
              onRegisterClick={() => setMode('register')}
            />
          ) : (
            <RegisterForm
              onSuccess={() => {}}
              onLoginClick={() => setMode('login')}
            />
          )}
        </div>
      </div>
    </div>
  );
}
