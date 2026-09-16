import { useState } from 'react';

interface RegisterPageProps {
  onSwitchToLogin: () => void;
}

function RegisterPage({ onSwitchToLogin }: RegisterPageProps) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = () => {
    console.log({ email, password });
  };

  return (
    <div className="auth-page">
      <div className='auth-card'>
        <h1 className='auth-title'>Create an account</h1>

        <div className="auth-field">
          <label htmlFor="email">Email</label>
          <input
            id="email"
            type="text"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Your email"
          />
        </div>

        <div className="auth-field">
          <label htmlFor="password">Password</label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Your password"
          />
        </div>
        <button className='auth-submit' onClick={handleSubmit}>Sign up</button>
        <p className="auth-switch">
          Already have an account ?{' '}
          <button onClick={onSwitchToLogin}>Log in</button>
        </p>
      </div>
    </div>
  );
}

export default RegisterPage;