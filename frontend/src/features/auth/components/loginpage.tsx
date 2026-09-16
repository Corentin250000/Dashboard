import { useState } from 'react';
import '../auth.css';

interface LoginPageProps {
  onSwitchToRegister: () => void;
}

function LoginPage({ onSwitchToRegister }: LoginPageProps) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = () => {
    console.log({ email, password });
  };

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1 className="auth-title">Log in</h1>

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

        <button className="auth-submit" onClick={handleSubmit}>
          Log in
        </button>

        <p className="auth-switch">
          Don't have an account ?{' '}
          <button onClick={onSwitchToRegister}>Create one</button>
        </p>
      </div>
    </div>
  );
}

export default LoginPage;