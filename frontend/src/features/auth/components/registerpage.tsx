import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import '../auth.css';

function RegisterPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const navigate = useNavigate();

  const handleSubmit = () => {
    setErrorMessage('');
    if (!email || !password){
      setErrorMessage('Please fill in all fields');
      return;
    }

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
    fetch(`${apiUrl}/auth/login`, {
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
    .then(data => {
      localStorage.setItem('token', data.token);
      navigate('/pending-verification');
    })
    .catch(error => {
      console.log('Connection error:', error);
      setErrorMessage(error);
    });
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
        
        {errorMessage && (
          <p className="auth-error">{errorMessage} </p>
        )}

        <p className="auth-switch">
          Already have an account ?{' '}
          <Link to='/login'> log in </Link>
        </p>
      </div>
    </div>
  );
}

export default RegisterPage;