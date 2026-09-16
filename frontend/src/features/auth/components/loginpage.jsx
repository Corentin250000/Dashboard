import { useState } from 'react';
import '../auth.css';

function LoginPage({ onSwitchToRegister }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = () => {
    console.log({ username, password });
  };

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1 className="auth-title">Log in</h1>

        <div className="auth-field">
          <label htmlFor="username">Username</label>
          <input
            id="username"
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Your username"
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