import { useEffect, useState } from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { isAuthenticated } from './session';

export default function ProtectedRoute() {
  const [authenticated, setAuthenticated] = useState<boolean | null>(null);
  const [error, setError] = useState('');

  const checkSession = () => {
    setAuthenticated(null);
    setError('');
    isAuthenticated().then(setAuthenticated).catch((cause: unknown) => {
      setError(cause instanceof Error ? cause.message : 'Unable to verify your session.');
    });
  };

  useEffect(() => { checkSession(); }, []);

  if (authenticated === null) {
    if (error) {
      return (
        <main role="alert" className="p-6 text-center text-(--text-primary)">
          <p>Unable to verify your session. Please try again.</p>
          <button type="button" onClick={checkSession} className="mt-3 underline">
            Retry
          </button>
        </main>
      );
    }
    return <p role="status">Loading…</p>;
  }
  if (!authenticated) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}
