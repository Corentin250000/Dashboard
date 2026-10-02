import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import LoginPage from './features/auth/pages/loginpage';
import RegisterPage from './features/auth/pages/registerpage';
import PendingVerification from './features/auth/pages/pendingverificationpage';
import VerifyEmailPage from './features/auth/pages/verifyemailpage';
import SubscriptionPage from './features/subscription/pages/subspage';
import Dashboard from './features/dashboard/pages/dashboardPage'
import ProtectedRoute from './features/auth/components/protectedRoute';
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
        <Route path="/dashboard" element={<Dashboard />} />

        <Route element={<ProtectedRoute />}>
          <Route path="/subscription" element={<SubscriptionPage />} />
        </Route>

      </Routes>
    </BrowserRouter>
  );
}

export default App
