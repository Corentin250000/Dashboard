import { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { api } from '../../shared/api';
import AuthCard from '../components/authCard';
 
type ConfirmStatus =
  | 'loading'
  | 'confirmed'
  | 'already_confirmed'
  | 'expired'
  | 'invalid'
  | 'error';
 
const MESSAGES: Record<ConfirmStatus, string> = {
  loading: 'Verifying your email…',
  confirmed: 'Your email has been confirmed! Redirecting…',
  already_confirmed: 'Your email was already confirmed. Redirecting…',
  expired: 'This confirmation link has expired.',
  invalid: 'This confirmation link is invalid.',
  error: 'Something went wrong. Please try again later.',
};
 
function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const [status, setStatus] = useState<ConfirmStatus>('loading');
  const navigate = useNavigate();
 
  useEffect(() => {
    const token = searchParams.get('token');
    let redirectTimer: ReturnType<typeof setTimeout>;
 
    if (!token) {
      setStatus('invalid');
      return;
    }
 
    api.confirmEmail(token)
      .then((data) => {
        setStatus(data.status);

        if (data.status === 'confirmed' || data.status === 'already_confirmed') {
          redirectTimer = setTimeout(() => navigate('/dashboard'), 5000);
        }
      })
      .catch(() => {
        setStatus('error');
      });

        return () => clearTimeout(redirectTimer);
    }, [searchParams, navigate]);
 
  const canResend = status === 'expired' || status === 'invalid';
 
  return (
    <AuthCard title="Email verification">
      <div className="mb-4 flex flex-col gap-1.5 text-[13px] text-[var(--text-secondary)]">
        <p>{MESSAGES[status]}</p>
      </div>

      {canResend && (
          <p className="mt-5 text-center text-sm text-[var(--text-secondary)]">
            Need a new link?{' '}
            <Link className="text-[var(--accent)] no-underline hover:underline" to="/pending-verification">
              Resend email
            </Link>
          </p>
        )}
    </AuthCard>
  );
}
 
export default VerifyEmailPage;
