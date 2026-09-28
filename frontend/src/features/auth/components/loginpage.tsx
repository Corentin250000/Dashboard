import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { api } from '../../shared/api';
import '../auth.css';


function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const navigate = useNavigate();

  const handleSubmit = () => {
    setErrorMessage('');

    if (!email || !password) {
      setErrorMessage('Please fill in all fields');
      return;
    }

    api.login(email, password)
      .then(data => {
        localStorage.setItem('token', data.token);
        navigate('/dashboard');
      })
      .catch(error => setErrorMessage(error.message))
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
          <Link className='auth-link' to="/register">Sign up</Link>
        </p>
      </div>
    </div>
  );
}

export default LoginPage;
