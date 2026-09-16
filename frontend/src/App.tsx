import { useState } from 'react';
import './App.css'
import LoginPage from './features/auth/components/loginpage.jsx'
import RegisterPage from './features/auth/components/registerpage.jsx'

function App() {
  const [currentPage, setCurrentPage] = useState('login');

  const onSwitchToRegister = () => {
    setCurrentPage('register');
  };

  const onSwitchToLogin = () => {
    setCurrentPage('login');
  };

  return (
    <div>
      {currentPage === 'login' ? (
        <LoginPage onSwitchToRegister={onSwitchToRegister} />
      ) : (
        <RegisterPage onSwitchToLogin={onSwitchToLogin} />
      )}
    </div>
  )
}

export default App
