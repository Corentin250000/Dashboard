import { useState } from 'react';
import '../auth.css';

interface LoginPageProps {
  onSwitchToRegister: () => void;
}

function LoginPage({ onSwitchToRegister }: LoginPageProps) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorMessage, setErrorMessage] = useState('');

  const handleSubmit = () => {
    setErrorMessage('');

    if (!email || !password) {
      setErrorMessage('Please fill in all fields');
      return;
    }

    fetch("auth/login",{
      method: 'POST',
      headers: {'Content-type': 'application/json'},
      body: JSON.stringify({email, password})
    })
    .then(response => {
      if (!response.ok) {
        throw new Error('Error ' + response.status);
      }
      return response.json();
    })
    .catch(error => {
      console.log('Connection error:', error);
      setErrorMessage(error);
    });
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

        {errorMessage && (
          <p className="auth-error">{errorMessage}</p>
        )}

        <p className="auth-switch">
          Don't have an account ?{' '}
          <button onClick={onSwitchToRegister}>Create one</button>
        </p>
      </div>
    </div>
  );
}

export default LoginPage;