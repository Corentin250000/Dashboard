import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { api, ApiError, API_URL } from '../../shared/api';
import AuthCard from '../components/authCard';
import googleIcon from '../../../assets/icons/google.svg';
import githubIcon from '../../../assets/icons/github.svg';

function LoginPage() {
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

    api.login(email, password)
      .then(() => {
        navigate('/dashboard');
      })
     .catch((error) => {
      setErrorMessage(
        error instanceof ApiError && error.status >= 500
          ? 'Server unavailable, please try again later.'
          : error.message
  );
});
  };

  return (
    <AuthCard title="Log in">
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
              rounded-lg border border-(--border-input) bg-(--card)
              px-3 py-2.5 text-[15px] text-(--text-primary)
              outline-none transition
              focus-visible:border-(--accent)
              focus-visible:ring-2 focus-visible:ring-(--accent)
            "
          />
        </div>

        <div className="mb-4 flex flex-col gap-1.5">
          <label htmlFor="password" className="text-[13px] text-(--text-primary)">
            Password
          </label>
          <input
            id="password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Your password"
            className="
              rounded-lg border border-(--border-input) bg-(--card)
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
          Log in
        </button>

        {errorMessage && (
          <p role="alert" className="mt-3 text-[15px] font-medium text-(--accent-secondary)">
            {errorMessage}
          </p>
        )}
      </form>

      <div className="my-5 flex items-center gap-3 text-xs text-(--text-secondary)">
        <span className="h-px flex-1 bg-(--border)" />
        <span>or continue with</span>
        <span className="h-px flex-1 bg-(--border)" />
      </div>

      <div className="flex gap-3">
        <a
          href={`${API_URL}/auth/google`}
          className="flex flex-1 items-center justify-center gap-2 rounded-lg border border-(--border-input) bg-(--card) px-3 py-2.5 text-sm font-medium text-(--text-primary) transition hover:bg-(--accent-soft)"
          aria-label="Continue with Google"
        >
          <img src={googleIcon} alt="" className="h-5 w-5" aria-hidden="true" />
          Google
        </a>
        <a
          href={`${API_URL}/auth/github`}
          className="flex flex-1 items-center justify-center gap-2 rounded-lg border border-(--border-input) bg-(--card) px-3 py-2.5 text-sm font-medium text-(--text-primary) transition hover:bg-(--accent-soft)"
          aria-label="Continue with GitHub"
        >
          <img src={githubIcon} alt="" className="h-5 w-5" aria-hidden="true" />
          GitHub
        </a>
      </div>

      <p className="mt-5 text-center text-sm text-(--text-secondary)">
        Don't have an account?{' '}
        <Link className="text-(--accent) underline" to="/register">
          Sign up
        </Link>
      </p>
    </AuthCard>
  );
}

export default LoginPage;
