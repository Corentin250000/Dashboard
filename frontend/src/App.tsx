import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import LoginPage from './features/auth/components/loginpage'
import RegisterPage from './features/auth/components/registerpage'
import PendingVerification from './features/auth/components/pendingverificationpage'
import ConfirmAuthentication from './features/auth/components/confirmauthpage';
import './App.css'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/login" replace />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/pending-verification" element={<PendingVerification />} />
        <Route path="/confirm-authentication" element={<ConfirmAuthentication />} />
        {/* <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/subscription" element={<Subscription />} />
                <Route path="/link-unavailable" element={<LinkUnavailable />} /> */}
      </Routes>
    </BrowserRouter>
  );
}

export default App
