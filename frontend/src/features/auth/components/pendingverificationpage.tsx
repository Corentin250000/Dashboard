import { useState } from 'react';
import { api, ApiError } from '../../shared/api';
import '../auth.css';
 
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
    <div className="auth-page">
      <div className="auth-card">
        <h1 className="auth-title">Check your emails!</h1>
 
        <div className="auth-field">
          <p>A verification email has been sent to you.</p>
          <p>Check your inbox and don't forget to check your spam folder.</p>
        </div>
 
        <button className="auth-submit" onClick={handleResend} disabled={isSending}>
          {isSending ? 'Sending…' : 'Resend email'}
        </button>
 
        {message && (<p className={messageType === 'success' ? 'auth-message' : 'auth-error'}> {message} </p>)}
        </div>
    </div>
  );
}
 
export default PendingVerificationPage;
