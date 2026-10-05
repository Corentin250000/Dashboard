import { useState } from 'react';
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
      .then((data) => {
        if (data.status === 'already_confirmed') {
            setMessageType('success');
            setMessage('Your email is already confirmed. You can log in.');
        } else if (data.status === 'invalid') {
            setMessageType('error');
            setMessage('Unable to resend the email. Please try registering again.');
        } else {
            setMessageType('success');
            setMessage('A new verification email has been sent.');
        }
      })
      .catch((error: unknown) => {
        if (error instanceof ApiError && error.status === 429) {
            setMessageType('error');
            setMessage('Too many attempts. Please wait a few minutes before trying again.');
        } else {
            setMessageType('error');
            setMessage('Something went wrong. Please try again later.');
        }
      })
      .finally(() => setIsSending(false));
  };
  return (
    <AuthCard title="Check your emails!">

      <div className="mb-4 flex flex-col gap-1.5 text-[13px] text-[var(--text-secondary)]">
        <p>A verification email has been sent to you.</p>
        <p>
          Check your inbox and don't forget to check your spam folder.
        </p>
      </div>

      <button
        className="
          bg-[var(--accent)]
          mt-2
          w-full
          rounded-lg
          px-3
          py-3
          text-[var(--text-primary)]
          text-[15px]
          font-medium
          transition
          hover:opacity-90
          disabled:cursor-not-allowed
          disabled:opacity-50
        "
        onClick={handleResend}
        disabled={isSending}
      >
        {isSending ? "Sending…" : "Resend email"}
      </button>

      {message && (
        <p role="alert"
          className={
            messageType === "success"
              ? "text-[15px] font-medium text-[var(--text-secondary)]"
              : "text-[15px] font-medium text-[var(--accent-secondary)]"
          }
        >
          {message}
        </p>
      )}

    </AuthCard>
  );
}
 
export default PendingVerificationPage;
