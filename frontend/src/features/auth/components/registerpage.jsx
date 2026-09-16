import { useState } from 'react';

function RegisterPage({ onSwitchToLogin }) {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');

    const handleSubmit = () => {
        console.log({ username, password });
    };

  return (
    <div className="auth-page">
        <div className='auth-card'>
            <h1 className='auth-title'>Create an account</h1>
            
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
            <button className='auth-submit' onClick={handleSubmit}>Sign up</button>
            <p className="auth-switch">
                Already have an account ?{' '}
                <button onClick={onSwitchToLogin}>Log in</button>
            </p>
        </div>
    </div>

  )
}

export default RegisterPage
