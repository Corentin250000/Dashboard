import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import LoginPage from './features/auth/pages/loginpage';
import RegisterPage from './features/auth/pages/registerpage';
import PendingVerification from './features/auth/pages/pendingverificationpage';
import VerifyEmailPage from './features/auth/pages/verifyemailpage';
import SubscriptionPage from './features/subscription/components/subspage';
import Dashboard from './features/dashboard/dashboard'
import './App.css'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/login" replace />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/pending-verification" element={<PendingVerification />} />
        <Route path="/confirm" element={<VerifyEmailPage />} />
        <Route path="/subscription" element={<SubscriptionPage />} />
        <Route path="/dashboard" element={<Dashboard />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App
