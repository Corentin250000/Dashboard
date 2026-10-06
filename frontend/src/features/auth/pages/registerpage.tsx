import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { api, ApiError } from '../../shared/api';
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
      .then(() => {
        navigate('/pending-verification');
      })
      .catch((error: unknown) => {
  setErrorMessage(
    error instanceof ApiError && error.status >= 500
      ? 'Server unavailable, please try again later.'
      : error instanceof Error ? error.message : 'Something went wrong.'
  );
});
  };

  return (
    <AuthCard title="Sign up">
      <form onSubmit={handleSubmit}>
        <div className="mb-4 flex flex-col gap-1.5">
          <label htmlFor="email" className="text-[13px] text-(--text-primary)">
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
              rounded-lg border border-(--border) bg-(--card)
              px-3 py-2.5 text-[15px] text-(--text-primary)
              outline-none transition
              focus-visible:border-(--accent)
              focus-visible:ring-2 focus-visible:ring-(--accent)
            "
          />
        </div>

        <div className="mb-4 flex flex-col gap-1.5">
          <label htmlFor="password" className="text-[13px] text-(--text-primary)">
            Password (Min. 8 characters)
          </label>
          <input
            id="password"
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Your password"
            className="
              rounded-lg border border-(--border) bg-(--card)
              px-3 py-2.5 text-[15px] text-(--text-primary)
              outline-none transition
              focus-visible:border-(--accent)
              focus-visible:ring-2 focus-visible:ring-(--accent)
            "
          />
        </div>

        <button
          type="submit"
          className="
            mt-2 w-full rounded-lg bg-(--accent) px-3 py-3
            text-[15px] font-medium text-(--text-on-accent)
            transition hover:opacity-90
          "
        >
          Sign up
        </button>

        {errorMessage && (
          <p role="alert" className="mt-3 text-[15px] font-medium text-(--accent-secondary)">
            {errorMessage}
          </p>
        )}
      </form>

      <p className="mt-5 text-center text-sm text-(--text-secondary)">
        Already have an account?{' '}
        <Link className="text-(--accent) no-underline hover:underline" to="/login">
          Log in
        </Link>
      </p>
    </AuthCard>
  );
}

export default RegisterPage;
