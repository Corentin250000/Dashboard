import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { api } from '../../shared/api';
import AuthCard from '../components/authCard';

function RegisterPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const navigate = useNavigate();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage('');

    if (!email || !password) {
      setErrorMessage('Please fill in all fields');
      return;
    }

    api.register(email, password)
      .then(data => {
        localStorage.setItem('token', data.token);
        navigate('/pending-verification');
      })
      .catch(error => setErrorMessage(error.message));
  };

  return (
    <AuthCard title="Sign up">
      <form onSubmit={handleSubmit}>
        <div className="mb-4 flex flex-col gap-1.5">
          <label htmlFor="email" className="text-[13px] text-[var(--text-primary)]">
            Email
          </label>
          <input
            id="email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Your email"
            className="
              rounded-lg border border-[var(--border)] bg-[var(--card)]
              px-3 py-2.5 text-[15px] text-[var(--text-primary)]
              outline-none transition
              focus-visible:border-[var(--accent)]
              focus-visible:ring-2 focus-visible:ring-[var(--accent)]
            "
          />
        </div>

        <div className="mb-4 flex flex-col gap-1.5">
          <label htmlFor="password" className="text-[13px] text-[var(--text-primary)]">
            Password
          </label>
          <input
            id="password"
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Your password"
            className="
              rounded-lg border border-[var(--border)] bg-[var(--card)]
              px-3 py-2.5 text-[15px] text-[var(--text-primary)]
              outline-none transition
              focus-visible:border-[var(--accent)]
              focus-visible:ring-2 focus-visible:ring-[var(--accent)]
            "
          />
        </div>

        <button
          type="submit"
          className="
            mt-2 w-full rounded-lg bg-[var(--accent)] px-3 py-3
            text-[15px] font-medium text-[var(--text-on-accent)]
            transition hover:opacity-90
          "
        >
          Sign up
        </button>

        {errorMessage && (
          <p role="alert" className="mt-3 text-[15px] font-medium text-[var(--accent-secondary)]">
            {errorMessage}
          </p>
        )}
      </form>

      <p className="mt-5 text-center text-sm text-[var(--text-secondary)]">
        Already have an account?{' '}
        <Link className="text-[var(--accent)] no-underline hover:underline" to="/register">
          Log in
        </Link>
      </p>
    </AuthCard>
  );
}

export default RegisterPage;