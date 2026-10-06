import { useState } from 'react';
import { Link } from 'react-router-dom';
import { api, ApiError } from '../../shared/api';
import AuthCard from '../components/authCard';

function PendingVerificationPage() {
  const [message, setMessage] = useState('');
  const [isSending, setIsSending] = useState(false);
  const [messageType, setMessageType] = useState<'success' | 'error'>('error');

  const handleResend = () => {
    setMessage('');
    setIsSending(true);

    api.resendConfirmation()
      .then(() => {
        setMessageType('success');
        setMessage('A new verification email has been sent.');
      })
      .catch((error: unknown) => {
        setMessageType('error');
        if (error instanceof ApiError && error.status === 429) {
          setMessage('Too many attempts. Please wait a few minutes before trying again.');
        } else if (error instanceof ApiError && (error.status === 400 || error.status === 404)) {
          setMessage('Unable to resend the email. Please try registering again.');
        } else {
          setMessage('Something went wrong. Please try again later.');
        }
      })
      .finally(() => setIsSending(false));
  };

  return (
    <AuthCard title="Check your emails!">
      <div className="mb-4 flex flex-col gap-1.5 text-[13px] text-(--text-secondary)">
        <p>A verification email has been sent to you.</p>
        <p>Check your inbox and don't forget to check your spam folder.</p>
      </div>

      <button
        type="button"
        onClick={handleResend}
        disabled={isSending}
        className="
          mt-2 w-full rounded-lg bg-(--accent) px-3 py-3
          text-[15px] font-medium text-(--text-on-accent)
          transition hover:opacity-90
          disabled:cursor-not-allowed disabled:opacity-50
        "
      >
        {isSending ? 'Sending…' : 'Resend email'}
      </button>

      {message && (
        <p
          role="alert"
          className={`mt-3 text-[15px] font-medium ${
            messageType === 'success' ? 'text-(--text-secondary)' : 'text-(--accent-secondary)'
          }`}
        >
          {message}
        </p>
      )}

      <p className="mt-5 text-center text-sm text-(--text-secondary)">
        Already confirmed?{' '}
        <Link className="text-(--accent) no-underline hover:underline" to="/login">
          Log in
        </Link>
      </p>
    </AuthCard>
  );
}

export default PendingVerificationPage;
