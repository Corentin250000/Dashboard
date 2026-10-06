import { useEffect, useRef, useState } from 'react';
import { useSearchParams, useNavigate, Link } from 'react-router-dom';
import { api, ApiError } from '../../shared/api';
import AuthCard from '../components/authCard';

type ConfirmStatus = 'loading' | 'confirmed' | 'expired' | 'invalid' | 'error';
type ResendStatus = 'idle' | 'sending' | 'sent' | 'error';

const MESSAGES: Record<ConfirmStatus, string> = {
  loading: 'Verifying your email…',
  confirmed: 'Your email has been confirmed! Redirecting to dashboard…',
  expired: 'This confirmation link has expired.',
  invalid: 'This link is invalid or has already been used. If you already confirmed your account, you can log in.',
  error: 'Something went wrong. Please try again later.',
};

function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');
  const [status, setStatus] = useState<ConfirmStatus>(token ? 'loading' : 'invalid');
  const [resend, setResend] = useState<ResendStatus>('idle');
  const requested = useRef(false);
  const navigate = useNavigate();

  useEffect(() => {
    if (!token || requested.current) return;
    requested.current = true;

    api.confirmEmail(token)
      .then(() => setStatus('confirmed'))
      .catch((error: unknown) => {
        if (error instanceof ApiError && (error.status === 400 || error.status === 404)) {
          setStatus('invalid');
        } else if (error instanceof ApiError && error.status === 410) {
          setStatus('expired');
        } else {
          setStatus('error');
        }
      });
  }, [token]);

  useEffect(() => {
    if (status !== 'confirmed') return;
    const timer = setTimeout(() => navigate('/dashboard'), 3000);
    return () => clearTimeout(timer);
  }, [status, navigate]);

  const handleResend = () => {
  setResend('sending');
  api.resendConfirmation()
    .then(() => setResend('sent'))
    .catch(() => setResend('error'));
};

  return (
    <AuthCard title="Email verification">
      <p role="status" className="mb-4 text-[13px] text-(--text-secondary)">
        {MESSAGES[status]}
      </p>

      {(status === 'expired' || status === 'invalid' || status === 'error') && (
        <>
          <button
            type="button"
            onClick={handleResend}
            disabled={resend === 'sending' || resend === 'sent'}
            className="
              w-full rounded-lg bg-(--accent) px-3 py-3
              text-[15px] font-medium text-(--text-on-accent)
              transition hover:opacity-90
              disabled:cursor-not-allowed disabled:opacity-50
            "
          >
            {resend === 'sending' ? 'Sending…' : 'Send a new link'}
          </button>

          {resend === 'sent' && (
            <p role="status" className="mt-3 text-[15px] text-(--text-secondary)">
              A new verification email has been sent.
            </p>
          )}
          {resend === 'error' && (
            <p role="alert" className="mt-3 text-[15px] font-medium text-(--accent-secondary)">
              Unable to send a new link. Please try registering again.
            </p>
          )}
        </>
      )}

      {(status === 'invalid' || status === 'error') && (
        <p className="mt-5 text-center text-sm text-(--text-secondary)">
          <Link className="text-(--accent) no-underline hover:underline" to="/login">
            Log in
          </Link>
          {' · '}
          <Link className="text-(--accent) no-underline hover:underline" to="/register">
            Sign up
          </Link>
        </p>
      )}
    </AuthCard>
  );
}

export default VerifyEmailPage;
